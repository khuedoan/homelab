FROM alpine:3.19

# Install base packages
RUN apk add --no-cache \
    bash \
    curl \
    docker \
    git \
    make \
    python3 \
    py3-pip \
    sudo \
    unzip \
    openssh-client \
    wget \
    openssl \
    # Pre-installed Python packages
    py3-virtualenv \
    py3-setuptools

# Create a non-root user and add to sudoers
RUN addgroup -S vscode && \
    adduser -S vscode -G vscode -s /bin/bash && \
    mkdir -p /etc/sudoers.d && \
    echo "vscode ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/vscode && \
    chmod 0440 /etc/sudoers.d/vscode

# Set up a virtual environment for Python packages
RUN python3 -m venv /opt/venv && \
    /opt/venv/bin/pip install --no-cache-dir \
    ansible \
    ansible-lint \
    yamllint \
    pre-commit

# Add the virtual environment bin to the PATH
ENV PATH="/opt/venv/bin:$PATH"

# Install kubectl
RUN curl -LO "https://dl.k8s.io/release/$(curl -Ls https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && \
    chmod +x kubectl && \
    mv kubectl /usr/local/bin/

# Install k9s
RUN curl -sL https://github.com/derailed/k9s/releases/download/v0.31.0/k9s_Linux_amd64.tar.gz | tar -xz -C /usr/local/bin k9s && \
    chmod +x /usr/local/bin/k9s

# Install Helm
RUN curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 && \
    chmod +x get_helm.sh && \
    VERIFY_CHECKSUM=false ./get_helm.sh && \
    rm get_helm.sh

# Create directories for the user
USER vscode
RUN mkdir -p /home/vscode/.ssh && \
    chmod 700 /home/vscode/.ssh && \
    mkdir -p /home/vscode/.terraform.d && \
    chmod 700 /home/vscode/.terraform.d && \
    echo 'source /opt/venv/bin/activate' >> /home/vscode/.bashrc

WORKDIR /workspaces/homelab
