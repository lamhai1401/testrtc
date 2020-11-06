
## Requirement

- Ubuntu 20 TLS

## Download Vscode Server

```bash
wget -q https://github.com/cdr/code-server/releases/download/v3.5.0/code-server_3.5.0_amd64.deb

```

## Install vscode-server
```bash
dpkg -i code-server_3.5.0_amd64.deb

```

## Create a symbolic link
```bash
ln -s /usr/lib/code-server/code-server /usr/bin/code-server
```

## Create a folder for code-server, where it will store user data:
```bash
mkdir /var/lib/code-server
```

## Create vscode service:
```bash
nano /lib/systemd/system/code-server.service

[Unit]
Description=code-server
After=nginx.service

[Service]
Type=simple
Environment=PASSWORD=your_password
ExecStart=/usr/bin/code-server --bind-addr 127.0.0.1:8080 --user-data-dir /var/lib/code-server --auth password
Restart=always

[Install]
WantedBy=multi-user.target
```
### Change login password:
```bash
nano ~/.config/code-server/config.yaml
```

## Start code-server:
```bash
systemctl start code-server
systemctl enable code-server
```

## Generate SSL Letsencrypt:
```bash
apt install certbot -y
certbot certonly --standalone --agree-tos -m myemail@gmail.com -d domain.com
ls -lah /etc/letsencrypt/live/domain.com/
```

## Setup Nginx as a Reverse Proxy:
```bash
apt install nginx -y
nano /etc/nginx/sites-available/code-server.conf

server {
 listen 80;
 server_name domain.com;
 # enforce https
 return 301 https://$server_name:443$request_uri;
}

server {
 listen 443 ssl http2;
 server_name domain.com;

 ssl_certificate /etc/letsencrypt/live/domain.com/fullchain.pem;
 ssl_certificate_key /etc/letsencrypt/live/domain.com/privkey.pem;

 location / {
 proxy_pass http://127.0.0.1:8080/;
 proxy_set_header Host $host;
 proxy_set_header Upgrade $http_upgrade;
 proxy_set_header Connection upgrade;
 proxy_set_header Accept-Encoding gzip;
 }
}

ln -s /etc/nginx/sites-available/code-server.conf /etc/nginx/sites-enabled/
nginx -t
systemctl restart nginx
systemctl enable nginx
```

## Securing Your Domain:
```bash
ufw allow https
ufw reload
```