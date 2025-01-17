# Go Bindings for the NVIDIA Management Library (NVML)

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [How the bindings are generated](#how-the-bindings-are-generated)
- [Code Structure](#code-structure)
  - [Code defining the NVML API](#code-defining-the-nvml-api)
  - [Code to load `libnvidia-ml.so`](#code-to-load-libnvidia-mlso)
  - [Code to bridge the auto-generated and manual bindings](#code-to-bridge-the-auto-generated-and-manual-bindings)
  - [Manual wrappers around the auto-generated bindings from `c-for-go`](#manual-wrappers-around-the-auto-generated-bindings-from-c-for-go)
  - [Test code](#test-code)
- [Building and Testing](#building-and-testing)
- [Updating the Code](#updating-the-code)
  - [Update `nvml.h`](#update-nvmlh)
  - [Add new versioned APIs](#add-new-versioned-apis)
  - [Add manual wrappers](#add-manual-wrappers)
- [Releasing](#releasing)
- [Contributing](#contributing)

## Overview

This repository provides Go bindings for the [NVIDIA Management Library API
(NVML)](https://docs.nvidia.com/deploy/nvml-api/).

At present, these bindings are only supported on **Linux**.

These bindings are not a reimplementation of NVML in Go, but rather a set of
wrappers around the C API provided by `libnvidia-ml.so`. This library is part
of the standard [NVIDIA driver
distribution](https://www.nvidia.com/Download/index.aspx), and should be
available on any Linux system that has the NVIDIA driver installed.  The API is
designed to be backwards compatible, so the latest bindings should work with
any version of `libnvidia-ml.so` installed on your system.

**Note:** A working NVIDIA driver with `libnvidia-ml.so` is not required to
compile code that imports these bindings. However, you will get a runtime error
if `libnvidia-ml.so` is not available in your library path at runtime.

Please see the following link for documentation on the full NVML Go API:
<http://godoc.org/github.com/spheronFdn/nvml/pkg/nvml>

## Quick Start

All you need is a simple import and a call to `nvml.Init()` to start using
these bindings.

The code below shows an example of using these bindings to query all of the
GPUs on your system and print out their UUIDs.

```go
package main

import (
	"fmt"
	"log"

	"github.com/spheronFdn/nvml/pkg/nvml"
)

func main() {
	ret := nvml.Init()
	if ret != nvml.SUCCESS {
		log.Fatalf("Unable to initialize NVML: %v", nvml.ErrorString(ret))
	}
	defer func() {
		ret := nvml.Shutdown()
		if ret != nvml.SUCCESS {
			log.Fatalf("Unable to shutdown NVML: %v", nvml.ErrorString(ret))
		}
	}()

	count, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		log.Fatalf("Unable to get device count: %v", nvml.ErrorString(ret))
	}

	for i := 0; i < count; i++ {
		device, ret := nvml.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			log.Fatalf("Unable to get device at index %d: %v", i, nvml.ErrorString(ret))
		}

		uuid, ret := device.GetUUID()
		if ret != nvml.SUCCESS {
			log.Fatalf("Unable to get uuid of device at index %d: %v", i, nvml.ErrorString(ret))
		}

		fmt.Printf("%v\n", uuid)
	}
}
```

On my DGX workstation, this results in the following output:

```console
$ go run main.go
GPU-edfee158-11c1-52b8-0517-92f30e7fac88
GPU-f22fb098-d1b3-3806-2655-ba25f02229c1
GPU-f613f823-1032-b3ec-a876-50f2e35e6f9e
GPU-3109fa37-4445-73c7-b695-1b5a4d13f58e
GPU-e28a6529-288c-7ddf-8fea-68c4833cda70
GPU-a27fb382-bad2-c02a-95ba-f6a1da38e76c
GPU-f5bb8d07-ee19-1787-4d9a-a84c4ac6b086
GPU-1ba0ca0e-6d1d-d9db-07d8-c1c5a8c32814
```

## How the bindings are generated

This project leverages two core technologies:

1. Go's builtin support for `cgo` (<https://golang.org/cmd/cgo/>)
1. A third-party tool called `c-for-go` (<https://c.for-go.com/>)

Using these tools, we are able to generate a set of Go bindings for NVML, given
nothing more than a specific version of the `nvml.h` header file (which defines
the full NVML API). Most of the process to generate these bindings is
automated, but a few manual steps are required in order to make the generated
bindings more useful from an end user's perspective.

The basic flow to generate the bindings is therefore to:

1. Take the `nvml.h` file and pass it through `c-for-go`
1. Take each of the low-level Go bindings generated by `c-for-go` and wrap them
   in a more user-friendly API

As an example, consider the Go bindings generated for the
`nvmlDeviceGetAccountingPids()` API call below:

Original API in `nvml.h`:

```c
nvmlReturn_t nvmlDeviceGetAccountingPids(nvmlDevice_t device, unsigned int *count, unsigned int *pids);
```

Auto-generated Go bindings from `c-for-go`:

```go
func nvmlDeviceGetAccountingPids(Device Device, Count *uint32, Pids *uint32) Return {
	cDevice, _ := *(*C.nvmlDevice_t)(unsafe.Pointer(&Device)), cgoAllocsUnknown
	cCount, _ := (*C.uint)(unsafe.Pointer(Count)), cgoAllocsUnknown
	cPids, _ := (*C.uint)(unsafe.Pointer(Pids)), cgoAllocsUnknown
	__ret := C.nvmlDeviceGetAccountingPids(cDevice, cCount, cPids)
	__v := (Return)(__ret)
	return __v
}
```

Manual wrapper around the auto-generated bindings:

```go
package nvml

func DeviceGetAccountingPids(Device Device) ([]int, Return) {
	var Count uint32 = 1 // Will be reduced upon returning
	for {
		Pids := make([]uint32, Count)
		ret := nvmlDeviceGetAccountingPids(Device, &Count, &Pids[0])
		if ret == SUCCESS {
			return uint32SliceToIntSlice(Pids[:Count]), ret
		}
		if ret != ERROR_INSUFFICIENT_SIZE {
			return nil, ret
		}
		Count *= 2
	}
}

func (Device Device) GetAccountingPids() ([]int, Return) {
	return DeviceGetAccountingPids(Device)
}
```

This manual wrapper makes it so that users don't have to write the boiler-plate
code of figuring out the correct `count` to pass into the API while at the same
time growing the `Pids` array and turning into a slice. It would be used as
follows:

```go
device, _ := nvml.DeviceGetHandleByIndex(0)
pids, _ := device.GetAccountingPids()
...
```

This is actually one of the more complicated examples. Most of the
manual wrappers are very simple and look similar to the following:

Original API in `nvml.h`:

```c
nvmlReturn_t nvmlDeviceGetUUID(nvmlDevice_t device, char *uuid, unsigned int length);
```

Auto-generated Go bindings from `c-for-go`:

```go
func nvmlDeviceGetUUID(Device Device, Uuid *byte, Length uint32) Return {
	cDevice, _ := *(*C.nvmlDevice_t)(unsafe.Pointer(&Device)), cgoAllocsUnknown
	cUuid, _ := (*C.char)(unsafe.Pointer(Uuid)), cgoAllocsUnknown
	cLength, _ := (C.uint)(Length), cgoAllocsUnknown
	__ret := C.nvmlDeviceGetUUID(cDevice, cUuid, cLength)
	__v := (Return)(__ret)
	return __v
}
```

Manual wrapper around the auto-generated bindings:

```go
package nvml

func DeviceGetUUID(Device Device) (string, Return) {
	Uuid := make([]byte, DEVICE_UUID_BUFFER_SIZE)
	ret := nvmlDeviceGetUUID(Device, &Uuid[0], DEVICE_UUID_BUFFER_SIZE)
	return string(Uuid[:clen(Uuid)]), ret
}

func (Device Device) GetUUID() (string, Return) {
	return DeviceGetUUID(Device)
}
```

While it does take some effort to take the auto-generated bindings and manually
wrap them in the more user-friendly API, this only has to be done once per API
call and then never touched again. As such, as new release of NVML come out,
only the new API calls will need to be added.

The following section goes into the details of how the code is structured, and
what each file's purpose is.

## Code Structure

There are two top-level directories in this repository:

- `/gen`
- `/pkg`

The `/gen` directory is used to house any code used in the _generation_ of the
final Go bindings. The `/pkg` directory is used to house any static packages
associated with this project as well as the actual Go bindings once they have
been generated. The one exception is the code used to dynamically load the
`libnvidia-ml.so` code from a host system and attach the go bindings to it. This
package requires no generated code and is housed statically under `pkg/dl`.
Once the code under `gen/nvml` has passed through `c-for-go` and any manual
wrappers applied, the final generated bindings are placed under `pkg/nvml`.

In general, the code used to generate the NVML Go bindings can be broken into 4 logical parts:

1. Code defining the NVML API and how any auto-generated bindings should be produced
1. Code responsible for dynamically loading `libnvidia-ml.so` on a host system and hooking it up to the bindings
1. Code bridging the gap between any auto-generated bindings and the manual wrappers around them
1. The manual wrappers themselves
1. Test code

Each of these parts is discussed in detail below, along with the files associated with them.

### Code defining the NVML API

The following files aid in defining the NVML API and how any auto-generated bindings should be produced from it.

- `gen/nvml/nvml.h`
- `gen/nvml/nvml.yml`

The `nvml.h` file is a direct copy of `nvml.h` from the NVIDIA driver.  Since
the NVML API is guaranteed to be backwards compatible, we should strive to keep
this always up to date with the latest.

**Note:** The make process modifies `nvml.h` in that it translates any opaque
types defined by `nvml.h` into something more recognizable by `cgo`.

For example:

```diff
-typedef struct nvmlDevice_st* nvmlDevice_t;
+typedef struct
+{
+   struct nvmlDevice_st* handle;
+} nvmlDevice_t;
```

The two statements are semantically equivalent in terms of how they are laid
out in memory, but `cgo` will only generate a unique type for `nvmlDevice_t`
when expressed as the latter. When building the bindings we first update
`nvml.h` using `sed`, and then run `c-for-go` over it.

Finally, the `nvml.yml` file is the input file to `c-for-go` that tells it how
to parse `nvml.h` and auto-generate bindings for it. Please see the [`c-for-go`
wiki](https://github.com/xlab/c-for-go/wiki) for more information about the
contents of this file and how it works.

### Code to load `libnvidia-ml.so`

The code under `pkg/dl` is responsible for dynamically loading the
`libnvidia-ml.so` binary from a host system and connecting the go bindings to
it. This happens under the hood whenever a user makes an `nvml.Init()` call. It
is transparent to the end user, and should work without any further
user-intervention.

Depending on the version of `libnvidia-ml.so` that is found, certain
_versioned_ symbols need to be updated.  At the time of this writing, these
symbols include the following (as defined in `nvml.h`):

```c
#ifndef NVML_NO_UNVERSIONED_FUNC_DEFS
    #define nvmlInit                                nvmlInit_v2
    #define nvmlDeviceGetPciInfo                    nvmlDeviceGetPciInfo_v3
    #define nvmlDeviceGetCount                      nvmlDeviceGetCount_v2
    #define nvmlDeviceGetHandleByIndex              nvmlDeviceGetHandleByIndex_v2
    #define nvmlDeviceGetHandleByPciBusId           nvmlDeviceGetHandleByPciBusId_v2
    #define nvmlDeviceGetNvLinkRemotePciInfo        nvmlDeviceGetNvLinkRemotePciInfo_v2
    #define nvmlDeviceRemoveGpu                     nvmlDeviceRemoveGpu_v2
    #define nvmlDeviceGetGridLicensableFeatures     nvmlDeviceGetGridLicensableFeatures_v3
    #define nvmlEventSetWait                        nvmlEventSetWait_v2
    #define nvmlDeviceGetAttributes                 nvmlDeviceGetAttributes_v2
    #define nvmlDeviceGetComputeRunningProcesses    nvmlDeviceGetComputeRunningProcesses_v2
    #define nvmlDeviceGetGraphicsRunningProcesses   nvmlDeviceGetGraphicsRunningProcesses_v2
#endif // #ifndef NVML_NO_UNVERSIONED_FUNC_DEFS
```

The actual versions that these API calls are assigned to will depend on the
version of the NVIDIA driver (and hence the version of `libnvidia-ml.so` that
you have linked in). These updates happen in the `updateVersionedSymbols()`
function of `pkg/nvml/lib.go` as seen below.

```go
// Default all versioned APIs to v1 (to infer the types)
var nvmlInit = nvmlInit_v1
var nvmlDeviceGetPciInfo = nvmlDeviceGetPciInfo_v1
var nvmlDeviceGetCount = nvmlDeviceGetCount_v1
...

// updateVersionedSymbols checks for versioned symbols in the loaded dynamic library.
// If newer versioned symbols exist, these replace the default `v1` symbols initialized above.
// When new versioned symbols are added, these would have to be initialized above and have
// corresponding checks and subsequent assignments added below.
func (l *library) updateVersionedSymbols() {
	ret := l.Lookup("nvmlInit_v2")
	if ret == SUCCESS {
		nvmlInit = nvmlInit_v2
	}
	ret = l.Lookup("nvmlDeviceGetPciInfo_v2")
	if ret == SUCCESS {
		nvmlDeviceGetPciInfo = nvmlDeviceGetPciInfo_v2
	}
	ret = l.Lookup("nvmlDeviceGetPciInfo_v3")
	if ret == SUCCESS {
		nvmlDeviceGetPciInfo = nvmlDeviceGetPciInfo_v3
	}
	ret = l.Lookup("nvmlDeviceGetCount_v2")
	if ret == SUCCESS {
		nvmlDeviceGetCount = nvmlDeviceGetCount_v2
	}
	...
}
```

Whenever a new version of NVML comes out that either (1) adds a new versioned
API call, or (2) bumps the version of an existing API call -- we need to make
sure and update this function appropriately (as well as make the necessary
changes to `nvidia.yml` to ensure all `v1` symbols are imported appropriately).

### Code to bridge the auto-generated and manual bindings

The files below define a set of "glue" code between the auto-generated bindings
from `c-for-go` and the manual wrappers providing a more user-friendly API to
the end user.

- `pkg/nvml/cgo_helpers_atatic.go`
- `pkg/nvml/return.go`

The `cgo_helpers.go` file defines functions that help in dealing with the types
coming out of the C API and turning them into more usable Go types. It is
actually a stripped down version of the auto-generated `cgo_helpers.go` file
from `c-for-go` that we have whittled down to the bare essentials. We also
define a few of our own functions in here as well. For example, doing things
like finding the length of a `NULL` terminated string inside a byte slice
(`clen()`), and converting a `uint32` slice into an `int` slice
(`uint32SliceToIntSlice()`), etc.

The `return.go` file simply wraps the `Return` type created by `c-for-go`
(which is a go-ified version of the `nvmlReturn_t` type from C) and has it
implement the `Error` interface so it can be returned as a normal Go `error`
type if desired. The string returned as part of the error is the result of
calling `nvmlErrorString()` under the hood.

### Manual wrappers around the auto-generated bindings from `c-for-go`

The following files add manual wrappers around all of the auto-generated
bindings from `c-for-go`. Only these manual wrappers are expected as part of
the API for the package -- the auto-generated bindings are only available for
internal use.

- `pkg/nvml/init.go`
- `pkg/nvml/system.go`
- `pkg/nvml/event_set.go`
- `pkg/nvml/vgpu.go`
- `pkg/nvml/unit.go`
- `pkg/nvml/device.go`

These wrappers add boiler-plate code around the auto-generated bindings so that
the end-user doesn't have to do this themselves every time a call is made.

When appropriate, they also bind functions to the top-level `types` that are
defined (e.g. `Unit`, `Device`, `EventSet`, `Vgpu`, etc.) so that functions can
be called directly on instances of these types instead of using a call at the
package scope.

A few examples of a this can be seen below:

```go
// nvml.UnitGetDevices()
func UnitGetDevices(Unit Unit) ([]Device, Return) {
	var DeviceCount uint32 = 1 // Will be reduced upon returning
	for {
		Devices := make([]Device, DeviceCount)
		ret := nvmlUnitGetDevices(Unit, &DeviceCount, &Devices[0])
		if ret == SUCCESS {
			return Devices[:DeviceCount], ret
		}
		if ret != ERROR_INSUFFICIENT_SIZE {
			return nil, ret
		}
		DeviceCount *= 2
	}
}

func (Unit Unit) GetDevices() ([]Device, Return) {
	return UnitGetDevices(Unit)
}


// nvml.DeviceGetUUID()
func DeviceGetUUID(Device Device) (string, Return) {
	Uuid := make([]byte, DEVICE_UUID_BUFFER_SIZE)
	ret := nvmlDeviceGetUUID(Device, &Uuid[0], DEVICE_UUID_BUFFER_SIZE)
	return string(Uuid[:clen(Uuid)]), ret
}

func (Device Device) GetUUID() (string, Return) {
	return DeviceGetUUID(Device)
}


// nvml.EventSetWait()
func EventSetWait(Set EventSet, Timeoutms uint32) (EventData, Return) {
	var Data EventData
	ret := nvmlEventSetWait(Set, &Data, Timeoutms)
	return Data, ret
}

func (Set EventSet) Wait(Timeoutms uint32) (EventData, Return) {
	return EventSetWait(Set, Timeoutms)
}


// nvml.VgpuInstanceGetUUID()
func VgpuInstanceGetUUID(VgpuInstance VgpuInstance) (string, Return) {
	Uuid := make([]byte, DEVICE_UUID_BUFFER_SIZE)
	ret := nvmlVgpuInstanceGetUUID(VgpuInstance, &Uuid[0], DEVICE_UUID_BUFFER_SIZE)
	return string(Uuid[:clen(Uuid)]), ret
}

func (VgpuInstance VgpuInstance) GetUUID() (string, Return) {
	return VgpuInstanceGetUUID(VgpuInstance)
}
```

Whenever a new version of NVML comes out that adds new API calls, a new set of
manual wrappers will need to be added to keep the API up-to-date. Adding the
initial set of wrappers was very time consuming, but adding additional wrappers
should be straightforward so long as we keep good pace with each new release.

### Test code

At present, all test code is under the following file:

- `pkg/nvml/nvml_test.go`

The test coverage is fairly sparse and could be greatly improved.

## Building and Testing

Building and testing the bindings is fairly straight-forward. The only
prerequisite is a working installation of `c-for-go` from
<https://github.com/xlab/c-for-go>.

**Note**: Please check the `Makefile` for the specific version of `c-for-go` used.

Once this is available, just run the sequence below to build and test these
NVML Go bindings. The generated bindings will be placed under `go-nvml/pkg/nvml`.

```console
$ make
c-for-go -out pkg gen/nvml/nvml.yml
  processing gen/nvml/nvml.yml done.
cp gen/nvml/*.go pkg/nvml
cd pkg/nvml; \
    go tool cgo -godefs types.go > types_gen.go; \
    go fmt types_gen.go; \
cd -> /dev/null
types_gen.go
rm -rf pkg/nvml/types.go pkg/nvml/_obj

$ make test
cd pkg/nvml; \
    go test -v .; \
cd -> /dev/null
=== RUN   TestInit
    TestInit: nvml_test.go:26: Init: Success
    TestInit: nvml_test.go:33: Shutdown: Success
--- PASS: TestInit (0.06s)
=== RUN   TestSystem
    TestSystem: nvml_test.go:45: SystemGetDriverVersion: Success
    TestSystem: nvml_test.go:46:   version: 410.104
    TestSystem: nvml_test.go:53: SystemGetNVMLVersion: Success
    TestSystem: nvml_test.go:54:   version: 10.410.104
    TestSystem: nvml_test.go:61: SystemGetCudaDriverVersion: Success
    TestSystem: nvml_test.go:62:   version: 10000
    TestSystem: nvml_test.go:69: SystemGetCudaDriverVersion_v2: Success
    TestSystem: nvml_test.go:70:   version: 10000
    TestSystem: nvml_test.go:77: SystemGetProcessName: Success
    TestSystem: nvml_test.go:78:   name: /lib/systemd/s
    TestSystem: nvml_test.go:85: SystemGetHicVersion: Success
    TestSystem: nvml_test.go:86:   count: 0
    TestSystem: nvml_test.go:96: SystemGetTopologyGpuSet: Success
    TestSystem: nvml_test.go:97:   count: 4
    TestSystem: nvml_test.go:99:   device[0]: {0x7f5875284408}
    TestSystem: nvml_test.go:99:   device[1]: {0x7f5875298e90}
    TestSystem: nvml_test.go:99:   device[2]: {0x7f58752ad918}
    TestSystem: nvml_test.go:99:   device[3]: {0x7f58752c23a0}
--- PASS: TestSystem (0.10s)
=== RUN   TestUnit
    TestUnit: nvml_test.go:112: UnitGetCount: Success
    TestUnit: nvml_test.go:113:   count: 0
    TestUnit: nvml_test.go:117: Skipping test with no Units.
--- SKIP: TestUnit (0.06s)
=== RUN   TestEventSet
    TestEventSet: nvml_test.go:253: EventSetCreate: Success
    TestEventSet: nvml_test.go:254:   set: {0x2122f10}
    TestEventSet: nvml_test.go:261: EventSetWait: Timeout
    TestEventSet: nvml_test.go:262:   data: {{<nil>} 0 0 0 0}
    TestEventSet: nvml_test.go:269: EventSet.Wait: Timeout
    TestEventSet: nvml_test.go:270:   data: {{<nil>} 0 0 0 0}
    TestEventSet: nvml_test.go:277: EventSetFree: Success
    TestEventSet: nvml_test.go:285: EventSet.Free: Success
--- PASS: TestEventSet (0.06s)
PASS
ok  github.com/spheronFdn/nvml/pkg/nvml 0.283s
```

**Note:** A working NVIDIA driver with `libnvidia-ml.so` is not required to
compile code that imports these bindings. However, you will get a runtime error
if `libnvidia-ml.so` is not available in your library path at runtime.

## Updating the Code

The general steps to update the bindings to a newer version of the NVML API are as follows:

### Update `nvml.h`

Pull down the `nvml.h` containing the updated API and commit it back to `gen/nvml/nvml.h`. The `Makefile` contains a command:

```console
$ make update-nvml-h
Found 5 NVML packages:

No.  Version   Upload Time          Package
  1  11.5.50   2021-11-23-22:46:02  nvidia/cuda-nvml-dev/11.5.50/linux-64/cuda-nvml-dev-11.5.50-h511b398_0.tar.bz2
  2  11.4.120  2021-11-03-22:08:33  nvidia/cuda-nvml-dev/11.4.120/linux-64/cuda-nvml-dev-11.4.120-hb8c74d6_0.tar.bz2
  3  11.4.43   2021-09-08-00:10:30  nvidia/cuda-nvml-dev/11.4.43/linux-64/cuda-nvml-dev-11.4.43-he36855d_0.tar.bz2
  4  11.3.58   2021-09-08-00:36:34  nvidia/cuda-nvml-dev/11.3.58/linux-64/cuda-nvml-dev-11.3.58-hc25e488_0.tar.bz2
  5  11.3.58   2021-09-08-00:36:31  nvidia/cuda-nvml-dev/11.3.58/linux-64/cuda-nvml-dev-11.3.58-h70090ce_0.tar.bz2

Pick an NVML package to update ([1]-5): 1

NVML version: 11.5.50
Package: nvidia/cuda-nvml-dev/11.5.50/linux-64/cuda-nvml-dev-11.5.50-h511b398_0.tar.bz2

Updating nvml.h to 11.5.50 from https://api.anaconda.org/download/nvidia/cuda-nvml-dev/11.5.50/linux-64/cuda-nvml-dev-11.5.50-h511b398_0.tar.bz2 ...
Successfully updated nvml.h to 11.5.50.
```

that copies the file from the Anaconda package [`anaconda.org/nvidia/cuda-nvml-dev`](https://anaconda.org/nvidia/cuda-nvml-dev).
Available files can be found at <https://anaconda.org/nvidia/cuda-nvml-dev/files> (platform: `linux-64`).

Since `gen/nvml/nvml.h` is under version control, running:

```bash
git diff -w
```

(ignoring whitespace) will show us which new API calls there are.

### Add new versioned APIs

If there are changes to the versioned APIs (defined as in the `#ifndef NVML_NO_UNVERSIONED_FUNC_DEFS` block in `gen/nvml/nvml.h`) `nvml.yml` and `init.go` must be updated accordingly.

The modified versioned calls can be found bu running:

```bash
git diff -w gen/nvml/nvml.h | grep -E "^\+\s*#define.*?_v[^1]"
```

### Add manual wrappers

Write a set of manual wrappers around any new calls as described in one of the previous sections above.

The following command should show the API calls added in the update:

```bash
git diff -w gen/nvml/nvml.h | grep "+nvmlReturn_t DECLDIR nvml"
```

Note that these includes the new versions of existing calls -- which should already have been handled in the previous section. To exclude these run:

```bash
git diff -w gen/nvml/nvml.h | grep "+nvmlReturn_t DECLDIR nvml" | grep -vE "_v\d+\("
```

Of course this is just the general flow, and there may be more work to do if
new types are added, or a new API is created that does something outside the
scope of what has been done so far. These guidelines should be a good starting
point though.

Keep in mind, that all updates to the NVML bindings code should be made in the
`gen/` directory of the repository. Only when releasing new bindings will this
code be processed and pushed to the `pkg/` directory for release.

## Releasing

Once the code in `gen/` has been fully updated to support a particular version of NVML, a new release should be created.

As part of the release, two things need to happen:

1. A new set of bindings needs to be generated from the code under `gen/` and committed into `pkg/`
1. A tag with the appropriate NVML release needs to be added to the repo and pushed upstream.

An example of this workflow for the 11.0 release of NVML can be seen below:

```bash
# Commit the generated bindings back to main
git checkout main
make
git add -f pkg/
git commit -m "Add bindings for v11.0 of the NVML API"
git push origin

# Tag the repo with the version number and push it upstream
git tag v11.0
git push origin v11.0
```

If updates need to be made against a particular version (due to bugs in the
bindings code, for example), then we append a `-<revision>` number to the
version tag we push.

For example:

```bash
git checkout v11.0
git checkout -b bug-fixes-for-v11.0
... fix bugs and commit
git tag v11.0-1
git push v11.0-1
```

Since the NVML API is designed to be backwards compatible, we envision it being
rare to require such backports (because people can just use the latest bindings
instead of relying on a particular version). However, we may perform such
backports from time-to-time as deemed necessary (or upon request).

## Contributing

Please see the file [CONTRIBUTING.md](CONTRIBUTING.md) for details on how to
contribute to this project.
