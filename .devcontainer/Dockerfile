FROM debian:bookworm-slim

# Install needed packages
RUN apt-get update && apt-get install -y \
    curl \
    git \
    gnupg \
    lsb-release \
    make \
    python3-pip \
    ssh \
    sudo \
    unzip \
    vim \
    wget \
    xz-utils

# Install Docker
RUN curl -fsSL https://get.docker.com | sh

# Install Ansible and other Python tools
RUN pip3 install ansible ansible-lint yamllint opentofu

# Install pre-commit
RUN pip3 install pre-commit

# Install k9s - Kubernetes CLI To Manage Your Clusters In Style
RUN curl -sL https://github.com/derailed/k9s/releases/download/v0.31.0/k9s_Linux_amd64.tar.gz | tar -xz -C /usr/local/bin k9s

# Install kubectl
RUN curl -LO "https://dl.k8s.io/release/$(curl -Ls https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && \
    chmod +x kubectl && \
    mv kubectl /usr/local/bin/

# Install Helm
RUN curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 && \
    chmod +x get_helm.sh && \
    ./get_helm.sh && \
    rm get_helm.sh

# Install Nix
ENV USER=root
RUN curl -L https://nixos.org/nix/install | sh -s -- --daemon
RUN echo 'source /etc/profile.d/nix.sh' >> /root/.bashrc
RUN echo 'experimental-features = nix-command flakes' >> /etc/nix/nix.conf

# Install direnv for nix shell automation
RUN curl -sfL https://direnv.net/install.sh | bash

# Add direnv hook to bashrc
RUN echo 'eval "$(direnv hook bash)"' >> /root/.bashrc

# Create .ssh directory
RUN mkdir -p /root/.ssh && chmod 700 /root/.ssh

# Add a note about mounting the .ssh directory
RUN echo "Note: Don't forget to create a .ssh directory in your project root with your SSH keys if you haven't already." > /root/SSH_NOTE.txt

WORKDIR /workspaces/homelab

CMD ["bash"]
