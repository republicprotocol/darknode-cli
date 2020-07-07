install=$1

$install 
if [ "$?" -eq "0" ]; then
    exit 1
fi

# Source the profile for the shell
. ~/.bash_profile
. ~/.bashrc
. ~/.profile

# Check If darknode has been installed properly
if ! darknode --version;then
    exit 1
fi