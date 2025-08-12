#!/bin/bash

# VM startup script for online judge executor service
# This script sets up Docker and deploys the executor service on a GCP Compute Engine VM

set -e

apt-get update
apt-get install -y apt-transport-https ca-certificates curl gnupg lsb-release

curl -fsSL https://download.docker.com/linux/debian/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
echo "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/debian $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null

apt-get update
apt-get install -y docker-ce docker-ce-cli containerd.io

systemctl start docker
systemctl enable docker

useradd -m -s /bin/bash executor
usermod -aG docker executor

mkdir -p /opt/online-judge-executor
chown executor:executor /opt/online-judge-executor

cat > /etc/systemd/system/online-judge-executor.service << 'EOF'
[Unit]
Description=Online Judge Executor Service
After=docker.service
Requires=docker.service

[Service]
Type=simple
User=executor
Group=executor
WorkingDirectory=/opt/online-judge-executor
ExecStart=/usr/bin/docker run --rm --name online-judge-executor \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -e RABBITMQ_HOST=${RABBITMQ_HOST} \
  -e RABBITMQ_PORT=${RABBITMQ_PORT} \
  -e RABBITMQ_USER=${RABBITMQ_USER} \
  -e RABBITMQ_PASS=${RABBITMQ_PASS} \
  ${EXECUTOR_IMAGE}
ExecStop=/usr/bin/docker stop online-judge-executor
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

cat > /opt/online-judge-executor/.env.template << 'EOF'
RABBITMQ_HOST=
RABBITMQ_PORT=5672
RABBITMQ_USER=
RABBITMQ_PASS=
EXECUTOR_IMAGE=
EOF

chown executor:executor /opt/online-judge-executor/.env.template

systemctl daemon-reload
systemctl enable online-judge-executor

curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -
echo "deb https://packages.cloud.google.com/apt cloud-sdk main" | tee -a /etc/apt/sources.list.d/google-cloud-sdk.list
apt-get update
apt-get install -y google-cloud-cli

cat > /opt/online-judge-executor/deploy.sh << 'EOF'
#!/bin/bash

set -e

if [[ -z "$RABBITMQ_HOST" || -z "$RABBITMQ_USER" || -z "$RABBITMQ_PASS" || -z "$EXECUTOR_IMAGE" ]]; then
    echo "Error: Required environment variables not set"
    echo "Required: RABBITMQ_HOST, RABBITMQ_USER, RABBITMQ_PASS, EXECUTOR_IMAGE"
    exit 1
fi

systemctl stop online-judge-executor || true

docker pull "$EXECUTOR_IMAGE"

docker rm -f online-judge-executor || true

cat > /opt/online-judge-executor/.env << EOF_ENV
RABBITMQ_HOST=$RABBITMQ_HOST
RABBITMQ_PORT=${RABBITMQ_PORT:-5672}
RABBITMQ_USER=$RABBITMQ_USER
RABBITMQ_PASS=$RABBITMQ_PASS
EXECUTOR_IMAGE=$EXECUTOR_IMAGE
EOF_ENV

systemctl start online-judge-executor

sleep 5
systemctl status online-judge-executor

echo "Deployment completed successfully"
EOF

chmod +x /opt/online-judge-executor/deploy.sh
chown executor:executor /opt/online-judge-executor/deploy.sh

echo "VM setup completed. Executor service is ready for deployment."