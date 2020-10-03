# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
  config.vm.box = "centos/8"

  config.vm.provider "virtualbox" do |vb|
    vb.cpus   = 4
    vb.memory = 8192
  end

  config.vm.network :public_network,
    :dev => "virbr0",
    :mode => "bridge",
    :type => "bridge"

  config.disksize.size = '100GB'

  config.vm.provision "ansible" do |ansible|
    ansible.playbook = "playbook.yml"
  end
end
