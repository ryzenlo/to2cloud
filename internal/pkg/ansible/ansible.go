package ansible

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"os"
	"os/exec"
	"ryzenlo/to2cloud/configs"
	"ryzenlo/to2cloud/internal/pkg/log"
)

const TempFilePrefix = "ansible"

type Inventory struct {
	Name          string
	Host          string
	User          string
	SSHPrivateKey string
}

type PlayCmd struct {
	cmdExecutor      *exec.Cmd
	checkCmdExecutor *exec.Cmd
	fullCmd          string
	checkSyntaxCmd   string
	inventory        Inventory
	playBookContent  string
	extraVars        map[string]string
	tmpInventory     *os.File
	tmpSSHPrivate    *os.File
}

var ProxyCommandValueMap = map[string]string{
	"socks5":  "nc -X 5 -x %s:%s %%h %%p",
	"socks4":  "nc -X 4 -x %s:%s %%h %%p",
	"default": "nc -x %s:%s %%h %%p",
}

func NewPlayCmd(c *configs.Config, playbookName string, i Inventory, proxyConfig configs.ProxyConfig, extraVars map[string]string) (*PlayCmd, error) {
	playCmd := &PlayCmd{
		inventory:        i,
		cmdExecutor:      &exec.Cmd{},
		checkCmdExecutor: &exec.Cmd{},
		extraVars:        extraVars,
	}
	//
	var err error
	var rawPlayBook []byte

	playbookPath := fmt.Sprintf("%s/%s", c.Ansible.DirPath, playbookName)
	if _, err := os.Stat(playbookPath); errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("the playbook file [%s] does not exist", playbookName)
	}
	if rawPlayBook, err = ioutil.ReadFile(playbookPath); err != nil {
		return nil, fmt.Errorf("failed to read playbook file content,%s", err.Error())
	}
	playCmd.playBookContent = string(rawPlayBook)

	var cmdPath, bashPath string
	if cmdPath, err = GetPlayBookCmdPathName(); err != nil {
		return nil, err
	}

	if bashPath, err = GetBashCmdPathName(); err != nil {
		return nil, err
	}

	//tmp file for inventory
	if playCmd.tmpInventory, err = ioutil.TempFile("/tmp", fmt.Sprintf("%s_inventory_", TempFilePrefix)); err != nil {
		return nil, fmt.Errorf("cannot create temporary inventory file,%w", err)
	}
	ic := GenerateInventoryContent(i, proxyConfig)
	//
	if _, err = playCmd.tmpInventory.WriteString(ic); err != nil {
		return nil, fmt.Errorf("cannot write temporary inventory file,%w", err)
	}
	if err = playCmd.tmpInventory.Sync(); err != nil {
		return nil, fmt.Errorf("cannot write temporary inventory file,%w", err)
	}
	//tmp file for ssh private key
	if playCmd.tmpSSHPrivate, err = ioutil.TempFile("/tmp", fmt.Sprintf("%s_key_", TempFilePrefix)); err != nil {
		return nil, fmt.Errorf("cannot create temporary ssh private file,%w", err)
	}
	//
	if _, err = playCmd.tmpSSHPrivate.WriteString(i.SSHPrivateKey); err != nil {
		return nil, fmt.Errorf("cannot write temporary ssh private file,%w", err)
	}
	if err = playCmd.tmpSSHPrivate.Sync(); err != nil {
		return nil, fmt.Errorf("cannot write temporary ssh private file,%w", err)
	}
	//
	playCmd.cmdExecutor.Path = bashPath
	playCmd.checkCmdExecutor.Path = bashPath
	//
	playCmd.fullCmd = GeneratePlaybookCommand(
		cmdPath,
		playbookPath,
		playCmd.tmpInventory.Name(),
		playCmd.tmpSSHPrivate.Name(),
		false,
		false,
		playCmd.extraVars,
	)
	playCmd.cmdExecutor.Args = []string{bashPath, "-c", playCmd.fullCmd}
	playCmd.cmdExecutor.Env = []string{"ANSIBLE_HOST_KEY_CHECKING=False"}
	//
	playCmd.checkSyntaxCmd = GeneratePlaybookCommand(
		cmdPath,
		playbookPath,
		playCmd.tmpInventory.Name(),
		playCmd.tmpSSHPrivate.Name(),
		true,
		false,
		playCmd.extraVars,
	)
	playCmd.checkCmdExecutor.Args = []string{bashPath, "-c", playCmd.checkSyntaxCmd}
	playCmd.checkCmdExecutor.Env = []string{"ANSIBLE_HOST_KEY_CHECKING=False"}
	return playCmd, nil
}

func GenerateInventoryContent(i Inventory, proxyConfig configs.ProxyConfig) string {
	fileContent := fmt.Sprintf("%s ansible_host=%s ansible_user=%s", i.Name, i.Host, i.User)
	if proxyConfig.UseProxy {
		proxyCommand := ProxyCommandValueMap["default"]
		if _, ok := ProxyCommandValueMap[proxyConfig.Type]; ok {
			proxyCommand = ProxyCommandValueMap[proxyConfig.Type]
		}
		fileContent = fmt.Sprintf(
			"%s ansible_ssh_common_args='-o ProxyCommand=\"%s\"'",
			fileContent,
			proxyCommand,
		)
		fileContent = fmt.Sprintf(fileContent, proxyConfig.Host, proxyConfig.Port)
	}
	return fileContent
}

func GetPlayBookCmdPathName() (string, error) {
	return exec.LookPath("ansible-playbook")
}

func GetBashCmdPathName() (string, error) {
	return exec.LookPath("bash")
}

func GeneratePlaybookCommand(cmdPath, playbookPath, inventoryPath, privateKeyPath string, syntaxChecking, hostKeyChecking bool, extraVars map[string]string) string {
	fullCmd := fmt.Sprintf(
		"%s %s -i %s --private-key %s",
		cmdPath,
		playbookPath,
		inventoryPath,
		privateKeyPath,
	)
	extraVarsContent := formExtraVarsContent(extraVars)
	if extraVarsContent != "" {
		fullCmd = fmt.Sprintf("%s %s", fullCmd, extraVarsContent)
	}
	if syntaxChecking {
		fullCmd = fmt.Sprintf("%s %s", fullCmd, "--syntax-check")
	}
	if hostKeyChecking {
		fullCmd = fmt.Sprintf("%s %s", "ANSIBLE_HOST_KEY_CHECKING=False", fullCmd)
	}
	return fullCmd
}

func (cmd *PlayCmd) CheckPlaybookSyntax() error {
	if cmd.checkSyntaxCmd == "" {
		return fmt.Errorf("empty command for running ansible playbook")
	}
	var out bytes.Buffer
	cmd.checkCmdExecutor.Stdout = &out
	cmd.checkCmdExecutor.Stderr = &out
	err := cmd.checkCmdExecutor.Run()
	log.Logger.Debugf("check run playbook command: %s", cmd.checkCmdExecutor.String())
	if err != nil {
		log.Logger.Infof("check playbook syntax command failed: %s", out.String())
	}
	return err
}

func (cmd *PlayCmd) Run() (string, error) {
	if cmd.fullCmd == "" {
		return "", fmt.Errorf("empty command for running ansible playbook")
	}
	var out bytes.Buffer
	cmd.cmdExecutor.Stdout = &out
	cmd.cmdExecutor.Stderr = &out
	log.Logger.Debugf("run playbook command: %s", cmd.cmdExecutor.String())
	if err := cmd.cmdExecutor.Run(); err != nil {
		return "", fmt.Errorf("%w,%s", err, out.String())
	}
	return out.String(), nil
}

func (cmd *PlayCmd) GetFullCmd() string {
	return cmd.fullCmd
}

func (cmd *PlayCmd) GetInventoryContent() string {
	fc, _ := ioutil.ReadFile(cmd.tmpInventory.Name())
	return string(fc)
}

func (cmd *PlayCmd) GetKeyContent() string {
	fc, _ := ioutil.ReadFile(cmd.tmpSSHPrivate.Name())
	return string(fc)
}

func (cmd *PlayCmd) GetPlayBookContent() string {
	return cmd.playBookContent
}

func (cmd *PlayCmd) GetAnsibleExtraVariables() string {
	if cmd.extraVars == nil {
		return ""
	}
	raw, err := json.Marshal(cmd.extraVars)
	if err != nil {
		return ""
	}
	return string(raw)
}

func formExtraVarsContent(extraVars map[string]string) string {
	if extraVars == nil {
		return ""
	}
	raw, err := json.Marshal(extraVars)
	if err != nil {
		return ""
	}
	content := fmt.Sprintf("-e '%s'", string(raw))
	return content
}

func (cmd *PlayCmd) Clean() {
	if cmd.tmpInventory != nil {
		cmd.tmpInventory.Close()
		os.Remove(cmd.tmpInventory.Name())
	}
	if cmd.tmpSSHPrivate != nil {
		cmd.tmpSSHPrivate.Close()
		os.Remove(cmd.tmpSSHPrivate.Name())
	}
}
