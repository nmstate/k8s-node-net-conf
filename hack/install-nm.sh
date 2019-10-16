#!/bin/bash -e

function install_nm_on_node() {
    node=$1
    $SSH $node sudo -- yum install -y NetworkManager NetworkManager-ovs
    $SSH $node sudo -- systemctl daemon-reload
    $SSH $node sudo -- systemctl restart NetworkManager
    echo "Check NetworkManager is working fine on node $node"
    $SSH $node -- nmcli device show > /dev/null
}

if [[ "$KUBEVIRT_PROVIDER" =~  k8s- ]]; then
    echo 'Install NetworkManager on nodes'
    for node in $($KUBECTL get nodes --no-headers | awk '{print $1}'); do
        install_nm_on_node "$node"
    done
fi