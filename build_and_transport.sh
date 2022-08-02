# Build program with selected architecture ($ARCH) and transport the binary to the gateway
# Assumes that the gateway accepts SSH connections on $SSH_PORT and that $DEPLOY_PATH is writable.
ARCH=mips
GW_IP=192.168.0.1
SSH_PORT=22
USERNAME=root
DEPLOY_PATH=/tmp/  # Where binary should live in the gateway

export TARGET_ARCH=$ARCH
make clean && make && scp -O -P $SSH_PORT build/bin/linux_${ARCH}/gowifi-datacollector $USERNAME@$GW_IP:$DEPLOY_PATH
