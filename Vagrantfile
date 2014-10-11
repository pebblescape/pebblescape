# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
  config.vm.box = "trusty64"
  config.vm.box_url = "http://cloud-images.ubuntu.com/vagrant/trusty/current/trusty-server-cloudimg-amd64-vagrant-disk1.box"
  config.vm.provision "ansible" do |ansible|
    ansible.playbook = "../ansible/pebblescape.yml"
  end
  config.ssh.forward_agent = true
  config.vm.network "forwarded_port", guest: 5000, host: 5000
  config.vm.network "forwarded_port", guest: 2341, host: 2341

  config.vm.define :mike do |pm_config|
    pm_config.vm.host_name = "mike.local"
    pm_config.vm.network "private_network", ip: "10.10.10.25"
    pm_config.vm.synced_folder "../", "/pebblescape"
  end
  
  config.vm.provider :virtualbox do |v|
    v.customize ["modifyvm", :id, "--memory", 2048]
  end
  
  config.vm.provider :vmware_fusion do |v|
    config.vm.define :mike do |s|
      v.vmx["memsize"] = "2048"
      v.vmx["displayName"] = "mike"
    end
  end
end