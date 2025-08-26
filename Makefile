proto-gen:
	@protoc --go_out=protobuf/product --go-grpc_out=protobuf/product \
	        proto/product.proto
	@protoc --go_out=protobuf/user --go-grpc_out=protobuf/user \
	        proto/user.proto