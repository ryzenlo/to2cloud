# to2cloud

## Introduction
This is a project for personal use. Its main purpose is 
as follows:
- Creating a virtual machine in Vultr using Vultr API
- Setting up Nginx, configuring a static website, uploading static files with ansible and geerlingguy.nginx role
- Setting up Cloudflare CDN for the website using Cloudflare API.

## Prerequisite
- Vultr account with API Key generated, enough balance for bill
- Cloudflare account with API Key generated
- Domain that Cloudflare has the right to manage
- A local proxy client in case it is unable to reach the created virtual machine 
 
## How to run

### With Docker 

Build Docker image with the Dockerfile in the root directory   

```bash
docker build -t to2cloud:lastest . 
```
Run Docker image   

```bash
docker run --name to2cloudApp -it -p 9000:9000 to2cloud:lastest 
```

### Without Docker
__Prerequisites:__  
- Golang
- Ansible
- Sqlite3
- netcat-bsd

Create database and import tables and data  
```
touch sqlite/to2cloud.db && sqlite3 sqlite/to2cloud.db < sqlite/database.dump
```
Build and run   
```
go run cmd/web/main.go;
```

## How to use

### Import postman/to2cloud.postman.json into Postman   

Call UserLogin API to get JWT token for the other APIs  

```
# Call the API in your postman or using curl command
curl --location --request POST 'http://localhost:9000/login' \
--header 'Content-Type: application/json' \
--data-raw '{"username":"admin","userpwd":"123456"}'
```

Call AddVultrCloudProvider API to add your Vultr account  

```
curl --location --request POST 'http://localhost:9000/cloud_provider' \
--header 'Content-Type: application/json' \
--header 'JWT-Token: {$jwt-token}' \
--data-raw '{ \
    "type": "vps", \
    "name": "vultr", \
    "account": "ryzenlo.20220222@gmail.com", \
    "vultr_config": {"api_key": "your-api-key", \
        "ssh_key_id": "your-ssh-key-id", \
        "ssh_private_key":"Your private key" \
    } \
}'

```

Call AddCloudflareCloudProvider API to add your Cloudflare account  

```
curl --location --request POST 'http://localhost:9000/cloud_provider' \
--header 'Content-Type: application/json' \
--header 'JWT-Token: {$jwt-token}' \
--data-raw '{  \ 
    "type": "cdn",  \
    "name": "cloudflare",  \
    "account": "ryzenlo.20220222@gmail.com",  \
    "cloudflare_config": {  \
        "api_key": "your-ssh-key-id",  \
        "email": "ryzenlo.20220222@gmail.com"  \
    }  \
}'
```

Check Vultr and Cloudflare account are setup properly  

```
# First,call GetCloudProviders API to get already added cloud providers
curl --location --request GET 'http://localhost:9000/cloud_providers'

# Call CheckVultrIsSetupProperly API to check Vultr account is setup properly
curl --location --request GET 'http://localhost:9000/cloud_provider/{id}/vultr/check' \
--header 'JWT-Token: {$jwt-token}'
# Call CheckCloudflareIsSetupProperly API to check Cloudflare account is setup properly
curl --location --request GET 'http://localhost:9000/cloud_provider/{id}/cloudflare/check' \
--header 'JWT-Token: {$jwt-token}'
```

Call AddSSHKey API to add your ssh public key in relation to the ssh private key in AddVultrCloudProvider API  

```
curl --location --request POST 'http://localhost:9000/cloud_provider/2/vultr/sshkey' \
--header 'JWT-Token: {$jwt-token}' \
--header 'Content-Type: application/json' \
--data-raw '{ \
    "name": "to2cloud", \
    "ssh_key":"your ssh public key" \
}'

```

Call CreateVultrInstance API to add a instance in Vultr   

Call RunPlayBook API to run the ansible playbook on the newly created instance   

Call GetOpsLogs API to get the result of running the ansible playbook  

Call UpdateDnsRecords API to update dns record for the newly created instance   

## TODO 
- Web UI
- CLI
- Github Action  