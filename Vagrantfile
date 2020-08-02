# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/bionic64"

  config.vm.provider "virtualbox" do |vb|
    vb.cpus   = 4
    vb.memory = 8192
  end

  config.vm.network "public_network", ip: "192.168.1.50"

  config.disksize.size = '100GB'

  config.vm.synced_folder "./", "/home/vagrant/data"

  config.vm.provision "ansible" do |ansible|
    ansible.playbook = "playbook.yml"
  end
end
