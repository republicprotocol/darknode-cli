#  Deployment tests
name="testing-do-$1"

# Deploy a darknode using DO
darknode up -do -name $name --do-token $do_token --tags testing
if [ "$?" -ne "0" ]; then
    echo "failed to deploy darknode"
    exit 1
fi
if ! darknode list; then
    echo "failed to list node"
    exit 1
fi
if ! darknode down $name -f; then
    echo "failed to destroy node"
    exit 1
fi

# Return error when providing no name
darknode up --do
if [ "$?" -eq "0" ]; then
    echo "failed to pass test without providing a node name"
    exit 1
fi

# TODO : deploy node with specific region and instance type

# Return error when providing an empty name
darknode up --do --name ""
if [ "$?" -eq "0" ]; then
    echo "failed to pass test when providing an empty name"
    exit 1
fi

# Return error when providing an invalid network name
darknode up --do --name $name --do-token $do_token --network invalid-network
if [ "$?" -eq "0" ]; then
    echo "failed to pass test when providing an invalid network"
    exit 1
fi

# Return error when providing an invalid digital ocean token
darknode up --do --name $name --do-token invalid-token
if [ "$?" -eq "0" ]; then
    echo "failed to pass test when providing invalid credentials"
    exit 1
fi
rm -rf "$HOME/.darknode/darknodes/$name"

# Return error when providing an invalid digital ocean region
darknode up --do --name $name --do-token $do_token --do-region invalid-region
if [ "$?" -eq "0" ]; then
    echo "failed to pass test when providing invalid region"
    exit 1
fi

# Return error when providing an invalid digital ocean droplet type
darknode up --do --name $name --do-token $do_token --do-droplet invalid-droplet-type
if [ "$?" -eq "0" ]; then
    echo "failed to pass test when providing invalid instance"
    exit 1
fi
darknode down $name -f
echo "All DO tests passed!"