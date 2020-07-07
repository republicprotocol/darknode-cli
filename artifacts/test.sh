install=$1

$install 
if [ "$?" -eq "0" ]; then
    exit 1
fi

# Source the profile for the shell
source ~/.bash_profile
source ~/.bashrc
source ~/.profile

# Check If darknode has been installed properly
if ! darknode --version;then
    exit 1
fi