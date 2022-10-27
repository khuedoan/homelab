.POSIX:

WORKON_HOME ?= .venv
VENV_BASE ?= $(WORKON_HOME)/homelab
PYTHON := ${VENV_BASE}/bin/python3
PYTHON_BIN_DIR := ${VENV_BASE}/bin

SSH_KEY_FILE ?= ~/.ssh/id_ed25519
ANSIBLE_EXTRA_ARGS ?=

default: prepare cluster

$(SSH_KEY_FILE):
	ssh-keygen -t ed25519 -f "$@"

.PHONY: venv
venv:
	@if [ ! -d "$(VENV_BASE)" ]; then \
		python3 -m venv $(VENV_BASE); \
	fi

.PHONY: requirements
requirements:
requirements: venv
	@echo Install requirements
	@${PYTHON} -m pip install -r requirements.txt > /dev/null

.PHONY: requirements-ansible
requirements-ansible:
requirements-ansible: requirements
	@echo Install ansible-galaxy requirements
	@${PYTHON_BIN_DIR}/ansible-galaxy install -r requirements.yml > /dev/null
	@${PYTHON_BIN_DIR}/ansible-galaxy collection install -r requirements.yml > /dev/null

.PHONY: first-boot
first-boot:
first-boot: $(SSH_KEY_FILE) requirements-ansible
	${PYTHON_BIN_DIR}/ansible-playbook playbooks/install/prepare.yml \
		-e prepare_additional_disks_force_format=true \
		-e ansible_user=root \
		--ask-pass \
		$(ANSIBLE_EXTRA_ARGS) \
		--limit 'arm'
	${PYTHON_BIN_DIR}/ansible-playbook playbooks/install/prepare.yml \
		-e prepare_additional_disks_force_format=true \
		$(ANSIBLE_EXTRA_ARGS) \
		--limit 'all:!arm'

.PHONY: prepare
prepare:
prepare: $(SSH_KEY_FILE) requirements-ansible
	${PYTHON_BIN_DIR}/ansible-playbook \
		$(ANSIBLE_EXTRA_ARGS) \
		playbooks/install/prepare.yml

.PHONY: cluster
cluster:
cluster: $(SSH_KEY_FILE) requirements-ansible
	${PYTHON_BIN_DIR}/ansible-playbook \
		$(ANSIBLE_EXTRA_ARGS) \
		playbooks/install/cluster.yml

.PHONY: uninstall-k3s
uninstall-k3s:
uninstall-k3s: $(SSH_KEY_FILE) requirements-ansible
	${PYTHON_BIN_DIR}/ansible-playbook \
		$(ANSIBLE_EXTRA_ARGS) \
		playbooks/uninstall/k3s.yml

.PHONY: uninstall-longhorn
uninstall-longhorn:
uninstall-longhorn: $(SSH_KEY_FILE) requirements-ansible
	${PYTHON_BIN_DIR}/ansible-playbook \
		$(ANSIBLE_EXTRA_ARGS) \
		playbooks/uninstall/longhorn.yml

.PHONY: console
console:
console: $(SSH_KEY_FILE) requirements
	${PYTHON_BIN_DIR}/ansible-console \
		$(ANSIBLE_EXTRA_ARGS) \
		--become
