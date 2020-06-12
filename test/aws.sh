#  Deployment tests
name="testing-aws-$1"

# Deploy a darknode using AWS
darknode up --aws --name $name --aws-access-key  $aws_access_key --aws-secret-key $aws_secret_key --tags mainnet,testing
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

# Return error when not providing a provider name
darknode up
if [ "$?" -eq "0" ]; then
    echo "failed to pass test without providing a provider name"
    exit 1
fi

# Return error when not providing a node name
darknode up --aws
if [ "$?" -eq "0" ]; then
    echo "failed to pass test without providing a node name"
    exit 1
fi

# TODO : deploy node with specific region and instance type

# Return error when providing an empty name
darknode up --aws --name ""
if [ "$?" -eq "0" ]; then
    echo "failed to pass test when providing an empty name"
    exit 1
fi

# Return error when providing an invalid network name
darknode up --aws --name $name --network invalid-network
if [ "$?" -eq "0" ]; then
    echo "failed to pass test when providing an invalid network"
    exit 1
fi

# Return error when providing an invalid AWS credentials
darknode up --aws --name $name --aws-access-key key --aws-secret-key value
if [ "$?" -eq "0" ]; then
    echo "failed to pass test when providing invalid credentials"
    exit 1
fi
rm -rf "$HOME/.darknode/darknodes/$name"

# Return error when providing an invalid aws region
darknode up --aws --name $name --aws-access-key  $aws_access_key --aws-secret-key $aws_secret_key --aws-region invalid-region
if [ "$?" -eq "0" ]; then
    echo "failed to pass test when providing invalid region"
    exit 1
fi

# Return error when providing an invalid aws instance type
darknode up --aws --name $name --aws-access-key  $aws_access_key --aws-secret-key $aws_secret_key --aws-instance invalid-instance
if [ "$?" -eq "0" ]; then
    echo "failed to pass test when providing invalid instance"
    exit 1
fi
darknode down $name -f
echo "All AWS tests passed!"