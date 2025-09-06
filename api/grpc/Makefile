PROTOS := $(wildcard *.proto)
BASENAMES := $(basename $(PROTOS))

PB_GO := $(foreach b,$(BASENAMES),$(b)/$(b).pb.go)
GRPC_GO := $(foreach b,$(BASENAMES),$(b)/$(b)_grpc.pb.go)

PB_BIN := $(foreach b,$(BASENAMES),$(b).pb)


all: $(BASENAMES) $(PB_GO) $(GRPC_GO) $(PB_BIN)

$(BASENAMES):
	@mkdir -p $@

%.pb.go: ../%.proto
	protoc --go_out=$(@D) --go_opt=paths=source_relative $(@D).proto

%_grpc.pb.go: ../%.proto
	protoc --go-grpc_out=$(@D) --go-grpc_opt=paths=source_relative $(@D).proto


%.pb: %.proto
	protoc --descriptor_set_out=$@ $< 
