module github.com/TimLai666/graft/examples/kitchensink

go 1.25.0

require (
	github.com/TimLai666/graft v0.0.0
	github.com/TimLai666/graft/graftapp v0.0.0
	github.com/gogpu/gg v0.48.11
	github.com/gogpu/ui v0.1.33
)

require (
	github.com/coregx/signals v0.1.0 // indirect
	github.com/go-text/typesetting v0.3.4 // indirect
	github.com/go-webgpu/goffi v0.5.3 // indirect
	github.com/go-webgpu/webgpu v0.5.2 // indirect
	github.com/gogpu/gogpu v0.42.0 // indirect
	github.com/gogpu/gpucontext v0.21.0 // indirect
	github.com/gogpu/gputypes v0.5.0 // indirect
	github.com/gogpu/naga v0.17.15 // indirect
	github.com/gogpu/wgpu v0.30.1 // indirect
	golang.org/x/image v0.42.0 // indirect
	golang.org/x/sys v0.46.0 // indirect
	golang.org/x/text v0.38.0 // indirect
)

replace github.com/TimLai666/graft => ../..

replace github.com/TimLai666/graft/graftapp => ../../graftapp
