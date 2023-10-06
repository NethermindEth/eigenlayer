
#!/bin/bash
echo 'Installing Ubuntu Archive Tools'
git clone https://git.launchpad.net/ubuntu-archive-tools
sudo apt-get install ubuntu-dev-tools -y
cd ubuntu-archive-tools
echo 'Copying Packages'
python3 copy-package -y -b -p nethermindeth --ppa-name=eigenlayer -s jammy --to-suite=focal eigenlayer
python3 copy-package -y -b -p nethermindeth --ppa-name=eigenlayer -s jammy --to-suite=lunar eigenlayer
python3 copy-package -y -b -p nethermindeth --ppa-name=eigenlayer -s jammy --to-suite=bionic eigenlayer
cd ..
echo 'Cleanup'
sudo rm -rf ubuntu-archive-tools
