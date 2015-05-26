package vbox

/*
#cgo CFLAGS: -I third_party/VirtualBoxSDK/sdk/bindings/c/include
#cgo CFLAGS: -I third_party/VirtualBoxSDK/sdk/bindings/c/glue
#cgo LDFLAGS: -ldl -lpthread

#include <stdlib.h>
#include "c_wrappers/machine.c"
*/
import "C"  // cgo's virtual package

import (
  "errors"
  "fmt"
  "reflect"
  "unsafe"
)

// The description of a VirtualBox machine
type Machine struct {
  cmachine *C.IMachine
}

// GetName returns the machine's name.
// It returns a string and any error encountered.
func (machine *Machine) GetName() (string, error) {
  var cname *C.char
  result := C.GoVboxGetMachineName(machine.cmachine, &cname)
  if C.GoVboxFAILED(result) != 0 || cname == nil {
    return "", errors.New(
        fmt.Sprintf("Failed to get IMachine name: %x", result))
  }

  name := C.GoString(cname)
  C.GoVboxUtf8Free(cname)
  return name, nil
}

// GetOsTypeId returns a string used to identify the guest OS type.
// It returns a string and any error encountered.
func (machine *Machine) GetOsTypeId() (string, error) {
  var cosTypeId *C.char
  result := C.GoVboxGetMachineOSTypeId(machine.cmachine, &cosTypeId)
  if C.GoVboxFAILED(result) != 0 || cosTypeId == nil {
    return "", errors.New(
        fmt.Sprintf("Failed to get IMachine OS type ID: %x", result))
  }

  osTypeId := C.GoString(cosTypeId)
  C.GoVboxUtf8Free(cosTypeId)
  return osTypeId, nil
}

// GetSettingsFilePath returns the path of the machine's settings file.
// It returns a string and any error encountered.
func (machine *Machine) GetSettingsFilePath() (string, error) {
  var cpath *C.char
  result := C.GoVboxGetMachineSettingsFilePath(machine.cmachine, &cpath)
  if C.GoVboxFAILED(result) != 0 || cpath == nil {
    return "", errors.New(
        fmt.Sprintf("Failed to get IMachine settings file path: %x", result))
  }

  path := C.GoString(cpath)
  C.GoVboxUtf8Free(cpath)
  return path, nil
}


// GetSettingsModified asks VirtualBox if this machine has unsaved settings.
// It returns a boolean and any error encountered.
func (machine *Machine) GetSettingsModified() (bool, error) {
  var cmodified C.PRBool
  result := C.GoVboxGetMachineSettingsModified(machine.cmachine, &cmodified)
  if C.GoVboxFAILED(result) != 0 {
    return false, errors.New(
        fmt.Sprintf("Failed to get IMachine modified flag: %x", result))
  }
  return cmodified != 0, nil
}

// SaveSettings saves a machine's modified settings.
// A new machine must have its settings saved before it can be registered.
// It returns a boolean and any error encountered.
func (machine *Machine) SaveSettings() error {
  result := C.GoVboxMachineSaveSettings(machine.cmachine)
  if C.GoVboxFAILED(result) != 0 {
    return errors.New(
        fmt.Sprintf("Failed to save IMachine settings: %x", result))
  }
  return nil
}

// Register adds this to VirtualBox's list of registered machines.
// It returns any error encountered.
func (machine *Machine) Register() error {
  // NOTE: This is a rare case where the underlying VirtualBox API call doesn't
  //       match the Go object model precisely. Register() really feels like it
  //       should belong to Machine and not to VirtualBox, because it takes a
  //       Machine argument, and VirtualBox is a singleton.
  result := C.GoVboxRegisterMachine(cbox, machine.cmachine)
  if C.GoVboxFAILED(result) != 0 {
    return errors.New(fmt.Sprintf("Failed to register IMachine: %x", result))
  }
  return nil
}

// Unregister removes this from VirtualBox's list of registered machines.
// The returned slice of Medium instances is intended to be passed to
// DeleteConfig to get all the VM's files cleaned.
// It returns an array of detached Medium instances and any error encountered.
func (machine *Machine) Unregister(cleanupMode CleanupMode) ([]Medium, error) {
  var cmediaPtr **C.IMedium
  var mediaCount C.ULONG

  result := C.GoVboxMachineUnregister(machine.cmachine,
      C.PRUint32(cleanupMode), &cmediaPtr, &mediaCount)
  if C.GoVboxFAILED(result) != 0 || (cmediaPtr == nil && mediaCount != 0) {
    return nil, errors.New(
        fmt.Sprintf("Failed to unregister machine: %x", result))
  }

  sliceHeader := reflect.SliceHeader{
    Data: uintptr(unsafe.Pointer(cmediaPtr)),
    Len:  int(mediaCount),
    Cap:  int(mediaCount),
  }
  cmediaSlice := *(*[]*C.IMedium)(unsafe.Pointer(&sliceHeader))

  var media = make([]Medium, mediaCount)
  for i := range cmediaSlice {
    media[i] = Medium{cmediaSlice[i]}
  }

  C.GoVboxArrayOutFree(unsafe.Pointer(cmediaPtr))
  return media, nil
}

// Release frees up the associated VirtualBox data.
// After the call, this instance is invalid, and using it will cause errors.
// It returns any error encountered.
func (machine* Machine) Release() error {
  if machine.cmachine != nil {
    result := C.GoVboxIMachineRelease(machine.cmachine)
    if C.GoVboxFAILED(result) != 0 {
      return errors.New(fmt.Sprintf("Failed to release IMachine: %x", result))
    }
    machine.cmachine = nil
  }
  return nil
}


// CreateMachine creates a VirtualBox machine.
// The machine must be registered by calling Register before it shows up in the
// GetMachines list.
// It returns the created machine and any error encountered.
func CreateMachine(
    name string, osTypeId string, flags string) (Machine, error) {
  var machine Machine
  if err := Init(); err != nil {
    return machine, err
  }

  cname := C.CString(name)
  cosTypeId := C.CString(osTypeId)
  cflags := C.CString(flags)
  result := C.GoVboxCreateMachine(cbox, cname, cosTypeId, cflags,
      &machine.cmachine)
  C.free(unsafe.Pointer(cname))
  C.free(unsafe.Pointer(cosTypeId))
  C.free(unsafe.Pointer(cflags))

  if C.GoVboxFAILED(result) != 0 || machine.cmachine == nil {
    return machine, errors.New(
        fmt.Sprintf("Failed to create IMachine: %x", result))
  }
  return machine, nil
}

// GetMachines returns the machines known to VirtualBox.
// It returns a slice of Machine instances and any error encountered.
func GetMachines() ([]Machine, error) {
  if err := Init(); err != nil {
    return nil, err
  }

  var cmachinesPtr **C.IMachine
  var machineCount C.ULONG

  result := C.GoVboxGetMachines(cbox, &cmachinesPtr, &machineCount)
  if C.GoVboxFAILED(result) != 0 ||
      (cmachinesPtr == nil && machineCount != 0) {
    return nil, errors.New(
        fmt.Sprintf("Failed to get IMachine array: %x", result))
  }

  sliceHeader := reflect.SliceHeader{
    Data: uintptr(unsafe.Pointer(cmachinesPtr)),
    Len:  int(machineCount),
    Cap:  int(machineCount),
  }
  cmachinesSlice := *(*[]*C.IMachine)(unsafe.Pointer(&sliceHeader))

  var machines = make([]Machine, machineCount)
  for i := range cmachinesSlice {
    machines[i] = Machine{cmachinesSlice[i]}
  }

  C.GoVboxArrayOutFree(unsafe.Pointer(cmachinesPtr))
  return machines, nil
}

