// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"github.com/spheronFdn/go-nvml/pkg/nvml"
	"sync"
)

// Ensure, that GpmSample does implement nvml.GpmSample.
// If this is not the case, regenerate this file with moq.
var _ nvml.GpmSample = &GpmSample{}

// GpmSample is a mock implementation of nvml.GpmSample.
//
//	func TestSomethingThatUsesGpmSample(t *testing.T) {
//
//		// make and configure a mocked nvml.GpmSample
//		mockedGpmSample := &GpmSample{
//			FreeFunc: func() nvml.Return {
//				panic("mock out the Free method")
//			},
//			GetFunc: func(device nvml.Device) nvml.Return {
//				panic("mock out the Get method")
//			},
//			MigGetFunc: func(device nvml.Device, n int) nvml.Return {
//				panic("mock out the MigGet method")
//			},
//		}
//
//		// use mockedGpmSample in code that requires nvml.GpmSample
//		// and then make assertions.
//
//	}
type GpmSample struct {
	// FreeFunc mocks the Free method.
	FreeFunc func() nvml.Return

	// GetFunc mocks the Get method.
	GetFunc func(device nvml.Device) nvml.Return

	// MigGetFunc mocks the MigGet method.
	MigGetFunc func(device nvml.Device, n int) nvml.Return

	// calls tracks calls to the methods.
	calls struct {
		// Free holds details about calls to the Free method.
		Free []struct {
		}
		// Get holds details about calls to the Get method.
		Get []struct {
			// Device is the device argument value.
			Device nvml.Device
		}
		// MigGet holds details about calls to the MigGet method.
		MigGet []struct {
			// Device is the device argument value.
			Device nvml.Device
			// N is the n argument value.
			N int
		}
	}
	lockFree   sync.RWMutex
	lockGet    sync.RWMutex
	lockMigGet sync.RWMutex
}

// Free calls FreeFunc.
func (mock *GpmSample) Free() nvml.Return {
	if mock.FreeFunc == nil {
		panic("GpmSample.FreeFunc: method is nil but GpmSample.Free was just called")
	}
	callInfo := struct {
	}{}
	mock.lockFree.Lock()
	mock.calls.Free = append(mock.calls.Free, callInfo)
	mock.lockFree.Unlock()
	return mock.FreeFunc()
}

// FreeCalls gets all the calls that were made to Free.
// Check the length with:
//
//	len(mockedGpmSample.FreeCalls())
func (mock *GpmSample) FreeCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockFree.RLock()
	calls = mock.calls.Free
	mock.lockFree.RUnlock()
	return calls
}

// Get calls GetFunc.
func (mock *GpmSample) Get(device nvml.Device) nvml.Return {
	if mock.GetFunc == nil {
		panic("GpmSample.GetFunc: method is nil but GpmSample.Get was just called")
	}
	callInfo := struct {
		Device nvml.Device
	}{
		Device: device,
	}
	mock.lockGet.Lock()
	mock.calls.Get = append(mock.calls.Get, callInfo)
	mock.lockGet.Unlock()
	return mock.GetFunc(device)
}

// GetCalls gets all the calls that were made to Get.
// Check the length with:
//
//	len(mockedGpmSample.GetCalls())
func (mock *GpmSample) GetCalls() []struct {
	Device nvml.Device
} {
	var calls []struct {
		Device nvml.Device
	}
	mock.lockGet.RLock()
	calls = mock.calls.Get
	mock.lockGet.RUnlock()
	return calls
}

// MigGet calls MigGetFunc.
func (mock *GpmSample) MigGet(device nvml.Device, n int) nvml.Return {
	if mock.MigGetFunc == nil {
		panic("GpmSample.MigGetFunc: method is nil but GpmSample.MigGet was just called")
	}
	callInfo := struct {
		Device nvml.Device
		N      int
	}{
		Device: device,
		N:      n,
	}
	mock.lockMigGet.Lock()
	mock.calls.MigGet = append(mock.calls.MigGet, callInfo)
	mock.lockMigGet.Unlock()
	return mock.MigGetFunc(device, n)
}

// MigGetCalls gets all the calls that were made to MigGet.
// Check the length with:
//
//	len(mockedGpmSample.MigGetCalls())
func (mock *GpmSample) MigGetCalls() []struct {
	Device nvml.Device
	N      int
} {
	var calls []struct {
		Device nvml.Device
		N      int
	}
	mock.lockMigGet.RLock()
	calls = mock.calls.MigGet
	mock.lockMigGet.RUnlock()
	return calls
}
