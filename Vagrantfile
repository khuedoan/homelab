# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/bionic64"

  config.vm.provider "virtualbox" do |vb|
    vb.cpus   = 4
    vb.memory = 8192
  end

  # LXD port
  config.vm.network "forwarded_port", guest: 8443, host: 8443

  config.disksize.size = '100GB'

  config.vm.provision "ansible" do |ansible|
    ansible.playbook = "playbook.yml"
  end
end
