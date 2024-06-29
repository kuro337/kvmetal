package db

const BKP = `
  - sudo -u ubuntu sh -c 'RUNZSH=no sh -c "$(curl -fsSL https://raw.github.com/ohmyzsh/ohmyzsh/master/tools/install.sh)"'
  - sudo -u ubuntu git clone https://github.com/zsh-users/zsh-autosuggestions /home/ubuntu/.oh-my-zsh/custom/plugins/zsh-autosuggestions
  - sudo -u ubuntu git clone https://github.com/zsh-users/zsh-syntax-highlighting /home/ubuntu/.oh-my-zsh/custom/plugins/zsh-syntax-highlighting
  - echo 'source /home/ubuntu/.oh-my-zsh/custom/plugins/zsh-autosuggestions/zsh-autosuggestions.zsh' | sudo -u ubuntu tee -a /home/ubuntu/.zshrc
  - echo 'source /home/ubuntu/.oh-my-zsh/custom/plugins/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh' | sudo -u ubuntu tee -a /home/ubuntu/.zshrc
  - echo 'plugins=(git zsh-autosuggestions zsh-syntax-highlighting)' | sudo -u ubuntu tee -a /home/ubuntu/.zshrc
  - sudo sed -i 's/PasswordAuthentication no/PasswordAuthentication yes/' /etc/ssh/sshd_config
  - sudo systemctl restart sshd

`

const CLICKHOUSE_RUNCMD = `
  - sudo -u ubuntu sh -c 'RUNZSH=no sh -c "$(curl -fsSL https://raw.github.com/ohmyzsh/ohmyzsh/master/tools/install.sh)"'
  - sudo -u ubuntu git clone https://github.com/zsh-users/zsh-autosuggestions /home/ubuntu/.oh-my-zsh/custom/plugins/zsh-autosuggestions
  - sudo -u ubuntu git clone https://github.com/zsh-users/zsh-syntax-highlighting /home/ubuntu/.oh-my-zsh/custom/plugins/zsh-syntax-highlighting
  - echo 'source /home/ubuntu/.oh-my-zsh/custom/plugins/zsh-autosuggestions/zsh-autosuggestions.zsh' | sudo -u ubuntu tee -a /home/ubuntu/.zshrc
  - echo 'source /home/ubuntu/.oh-my-zsh/custom/plugins/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh' | sudo -u ubuntu tee -a /home/ubuntu/.zshrc
  - echo 'plugins=(git zsh-autosuggestions zsh-syntax-highlighting)' | sudo -u ubuntu tee -a /home/ubuntu/.zshrc
  - sudo sed -i 's/PasswordAuthentication no/PasswordAuthentication yes/' /etc/ssh/sshd_config
  - sudo systemctl restart sshd

  - sudo apt-get install -y apt-transport-https ca-certificates curl gnupg
  - curl -fsSL 'https://packages.clickhouse.com/rpm/lts/repodata/repomd.xml.key' | sudo gpg --dearmor -o /usr/share/keyrings/clickhouse-keyring.gpg
  - echo "deb [signed-by=/usr/share/keyrings/clickhouse-keyring.gpg] https://packages.clickhouse.com/deb stable main" | sudo tee /etc/apt/sources.list.d/clickhouse.list
  - sudo apt-get update
  - sudo apt-get install -y clickhouse-server clickhouse-client
  - sudo systemctl restart clickhouse-server
`
