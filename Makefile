PYTHON     := python3
VENV_DIR   := $(PWD)/.venv

ANSIBLE_LIBRARY := $(PWD)/ansible/plugins/modules
export ANSIBLE_LIBRARY


venv:
	$(PYTHON) -m venv $(VENV_DIR)

install: venv 
	source $(VENV_DIR)/bin/activate; \
		$(PYTHON) -m pip install -r requirements.txt

clean:
	rm -rf $(VENV_DIR)

##
## Build Ansible modules
##
## `make build`
##
MODULES        := sftpgo-user 

MODULE_TARGETS := $(addprefix $(ANSIBLE_LIBRARY)/,$(MODULES))
DOC_TARGETS    := $(addsuffix .yml,$(MODULE_TARGETS))

GO_CMD_DIR     := $(PWD)/cmd
GO_SOURCES     := $(shell find $(PWD) -name '*.go')

build: $(MODULE_TARGETS) $(DOC_TARGETS)

$(ANSIBLE_LIBRARY)/%: $(GO_CMD_DIR)/%/main.go $(GO_SOURCES)
	go build -o $@ $(GO_CMD_DIR)/$*

$(ANSIBLE_LIBRARY)/%.yml: $(GO_CMD_DIR)/%/doc.yml
	cp $< $@


## 
## Examples 
## 
## `make example EXAMPLE_MODULE=sftpgo-user`
## `make example-doc EXAMPLE_MODULE=sftpgo-user`
##
EXAMPLE_MODULE ?= sftpgo-user

example: 
	source $(VENV_DIR)/bin/activate && \
		examples/ansible-playbook.sh -v -i examples/inventory.ini examples/$(EXAMPLE_MODULE)/example.yaml

example-doc: 
	source $(VENV_DIR)/bin/activate && \
		ansible-doc $(EXAMPLE_MODULE)
