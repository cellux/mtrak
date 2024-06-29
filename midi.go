package main

import (
	"fmt"
	"github.com/xthexder/go-jack"
)

type MidiMessage = [3]byte

type MidiEngine struct {
	client  *jack.Client
	outPort *jack.Port
}

type MidiData = jack.MidiData

func (e *MidiEngine) Open(processCallback jack.ProcessCallback) error {
	client, status := jack.ClientOpen("mtrak", jack.NoStartServer)
	if status != 0 {
		return fmt.Errorf("jack::ClientOpen() failed: %s", jack.StrError(status))
	}
	outPort := client.PortRegister("midi_out", jack.DEFAULT_MIDI_TYPE, jack.PortIsOutput, 0)
	if outPort == nil {
		client.Close()
		return fmt.Errorf("jack::PortRegister() failed")
	}
	if status := client.SetProcessCallback(processCallback); status != 0 {
		client.Close()
		return fmt.Errorf("jack::SetProcessCallback() failed: %s", jack.StrError(status))
	}
	if status := client.Activate(); status != 0 {
		client.Close()
		return fmt.Errorf("jack::Activate() failed: %s", jack.StrError(status))
	}
	e.client = client
	e.outPort = outPort
	return nil
}

func (e *MidiEngine) Close() error {
	if e.client != nil {
		e.client.Close()
		e.client = nil
	}
	return nil
}

func MidiMessageLength(status byte) int {
	switch status >> 4 {
	case 0x8:
		return 3 // note off
	case 0x9:
		return 3 // note on
	case 0xA:
		return 3 // aftertouch
	case 0xB:
		return 3 // controller
	case 0xC:
		return 2 // program change
	case 0xD:
		return 2 // channel pressure
	case 0xE:
		return 3 // pitch wheel
	case 0xF:
		return 0 // sysex (unsupported)
	default:
		return 0
	}
}
