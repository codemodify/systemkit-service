Vagrant.configure("2") do |config|
	config.vm.box = "freebsd/FreeBSD-12.1-STABLE"
	config.disksize.size = '50GB'
	config.ssh.shell = "sh"

  	config.vm.provider :virtualbox do |vb|
		vb.customize ["modifyvm", :id, "--memory", "8192"]
		vb.customize ["modifyvm", :id, "--cpus", "4"]
		vb.customize ["modifyvm", :id, "--ioapic", "on"]
	end

end
