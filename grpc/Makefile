GENDIR := api
PROTO_FIELS := $(wildcard pb/*/*.proto)

define exec-command
$(1)

endef

.PHONY: gen
gen:
	rm -rf ./$(GENDIR)
	mkdir -p ./$(GENDIR)
	$(foreach file,$(PROTO_FIELS),$(call exec-command, protoc --go_out ./$(GENDIR) --go-grpc_out ./$(GENDIR) $(file)))
	$(foreach file,$(PROTO_FIELS),$(call exec-command, protoc --grpc-gateway_out ./$(GENDIR) --grpc-gateway_opt logtostderr=true $(file)))