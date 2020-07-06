package provider

var templateDO = `
provider "digitalocean" {
  token = "{{.Token}}"
}

resource "digitalocean_ssh_key" "darknode" {
   name       = "{{.Name}}"
   public_key = file("{{.PubKeyPath}}")
}

resource "digitalocean_droplet" "darknode" {
  provider    = digitalocean
  image       = "ubuntu-18-04-x64"
  name        = "{{.Name}}"
  region      = "{{.Region}}"
  size        = "{{.Size}}"
  monitoring  = true
  resize_disk = false

  ssh_keys = [
    digitalocean_ssh_key.darknode.id
  ]

  provisioner "remote-exec" {
	
	inline = [
      "set -x",
      "until sudo apt update; do sleep 4; done",
      "until sudo apt-get -y update; do sleep 4; done",
      "sudo adduser darknode --gecos \",,,\" --disabled-password",
      "sudo rsync --archive --chown=darknode:darknode ~/.ssh /home/darknode",
	  "curl -sSL https://repos.insights.digitalocean.com/install.sh | sudo bash",
      "until sudo apt-get install ufw; do sleep 4; done",
      "sudo ufw limit 22/tcp",
      "sudo ufw allow 18514/tcp", 
      "sudo ufw allow 18515/tcp", 
      "sudo ufw --force enable",
	]

    connection {
      host        = self.ipv4_address
      type        = "ssh"
      user        = "root"
      private_key = file("{{.PriKeyPath}}")
    }
  }

  provisioner "file" {

    source      = "{{.ConfigPath}}"
    destination = "$HOME/config.json"

    connection {
      host        = self.ipv4_address
      type        = "ssh"
      user        = "darknode"
      private_key = file("{{.PriKeyPath}}")
    }
  }

  provisioner "remote-exec" {
	
	inline = [
      "set -x",
	  "mkdir -p $HOME/.darknode/bin",
      "mkdir -p $HOME/.config/systemd/user",
      "mv $HOME/config.json $HOME/.darknode/config.json",
	  "curl -sL https://www.github.com/renproject/darknode-release/releases/latest/download/darknode > ~/.darknode/bin/darknode",
	  "chmod +x ~/.darknode/bin/darknode",
      "echo {{.LatestVersion}} > ~/.darknode/version",
	  <<EOT
	  echo "{{.ServiceFile}}" > ~/.config/systemd/user/darknode.service
      EOT
      ,
	  "loginctl enable-linger darknode",
      "systemctl --user enable darknode.service",
      "systemctl --user start darknode.service",
	]

    connection {
      host        = self.ipv4_address
      type        = "ssh"
      user        = "darknode"
      private_key = file("{{.PriKeyPath}}")
    }
  }
}

output "provider" {
  value = "do"
}

output "ip" {
  value = digitalocean_droplet.darknode.ipv4_address
}`
