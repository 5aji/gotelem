
package skylab

import (
	"testing"
	"encoding/json"
)


func TestMarshalUnmarshalBmsMeasurement(t *testing.T) {
	v := &BmsMeasurement{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONBmsMeasurement(t *testing.T) {

	v := &BmsMeasurement{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *BmsMeasurement:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalBatteryStatus(t *testing.T) {
	v := &BatteryStatus{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONBatteryStatus(t *testing.T) {

	v := &BatteryStatus{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *BatteryStatus:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalBmsKillReason(t *testing.T) {
	v := &BmsKillReason{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONBmsKillReason(t *testing.T) {

	v := &BmsKillReason{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *BmsKillReason:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalBmsModuleMinMax(t *testing.T) {
	v := &BmsModuleMinMax{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONBmsModuleMinMax(t *testing.T) {

	v := &BmsModuleMinMax{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *BmsModuleMinMax:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalBmsSoc(t *testing.T) {
	v := &BmsSoc{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONBmsSoc(t *testing.T) {

	v := &BmsSoc{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *BmsSoc:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalBmsCapacity(t *testing.T) {
	v := &BmsCapacity{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONBmsCapacity(t *testing.T) {

	v := &BmsCapacity{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *BmsCapacity:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalBmsCurrentlimit(t *testing.T) {
	v := &BmsCurrentlimit{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONBmsCurrentlimit(t *testing.T) {

	v := &BmsCurrentlimit{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *BmsCurrentlimit:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalBmsFanInfo(t *testing.T) {
	v := &BmsFanInfo{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONBmsFanInfo(t *testing.T) {

	v := &BmsFanInfo{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *BmsFanInfo:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalBmsSetMinFanSpeed(t *testing.T) {
	v := &BmsSetMinFanSpeed{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONBmsSetMinFanSpeed(t *testing.T) {

	v := &BmsSetMinFanSpeed{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *BmsSetMinFanSpeed:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalBmsModule(t *testing.T) {
	v := &BmsModule{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONBmsModule(t *testing.T) {

	v := &BmsModule{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *BmsModule:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalBmsChargerResponse(t *testing.T) {
	v := &BmsChargerResponse{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONBmsChargerResponse(t *testing.T) {

	v := &BmsChargerResponse{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *BmsChargerResponse:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalDashboardPedalPercentages(t *testing.T) {
	v := &DashboardPedalPercentages{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONDashboardPedalPercentages(t *testing.T) {

	v := &DashboardPedalPercentages{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *DashboardPedalPercentages:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalCarState(t *testing.T) {
	v := &CarState{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONCarState(t *testing.T) {

	v := &CarState{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *CarState:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalDashboardPedalFault(t *testing.T) {
	v := &DashboardPedalFault{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONDashboardPedalFault(t *testing.T) {

	v := &DashboardPedalFault{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *DashboardPedalFault:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalDashboardSystemTimeoutTest(t *testing.T) {
	v := &DashboardSystemTimeoutTest{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONDashboardSystemTimeoutTest(t *testing.T) {

	v := &DashboardSystemTimeoutTest{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *DashboardSystemTimeoutTest:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalCarSpeed(t *testing.T) {
	v := &CarSpeed{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONCarSpeed(t *testing.T) {

	v := &CarSpeed{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *CarSpeed:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalFlightComputerLvBoardDisconnectCounts(t *testing.T) {
	v := &FlightComputerLvBoardDisconnectCounts{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONFlightComputerLvBoardDisconnectCounts(t *testing.T) {

	v := &FlightComputerLvBoardDisconnectCounts{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *FlightComputerLvBoardDisconnectCounts:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalFlightComputerHvBoardDisconnectCounts(t *testing.T) {
	v := &FlightComputerHvBoardDisconnectCounts{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONFlightComputerHvBoardDisconnectCounts(t *testing.T) {

	v := &FlightComputerHvBoardDisconnectCounts{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *FlightComputerHvBoardDisconnectCounts:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalFlightComputerInternalState(t *testing.T) {
	v := &FlightComputerInternalState{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONFlightComputerInternalState(t *testing.T) {

	v := &FlightComputerInternalState{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *FlightComputerInternalState:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalPowerToDrive(t *testing.T) {
	v := &PowerToDrive{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONPowerToDrive(t *testing.T) {

	v := &PowerToDrive{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *PowerToDrive:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalArrayPower(t *testing.T) {
	v := &ArrayPower{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONArrayPower(t *testing.T) {

	v := &ArrayPower{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *ArrayPower:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalVisionTurnSignalsCommand(t *testing.T) {
	v := &VisionTurnSignalsCommand{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONVisionTurnSignalsCommand(t *testing.T) {

	v := &VisionTurnSignalsCommand{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *VisionTurnSignalsCommand:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalVisionBrakeLightsCommand(t *testing.T) {
	v := &VisionBrakeLightsCommand{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONVisionBrakeLightsCommand(t *testing.T) {

	v := &VisionBrakeLightsCommand{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *VisionBrakeLightsCommand:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalVisionHeadlightsCommand(t *testing.T) {
	v := &VisionHeadlightsCommand{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONVisionHeadlightsCommand(t *testing.T) {

	v := &VisionHeadlightsCommand{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *VisionHeadlightsCommand:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalVisionHornCommand(t *testing.T) {
	v := &VisionHornCommand{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONVisionHornCommand(t *testing.T) {

	v := &VisionHornCommand{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *VisionHornCommand:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalVisionArrayLatchesCommand(t *testing.T) {
	v := &VisionArrayLatchesCommand{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONVisionArrayLatchesCommand(t *testing.T) {

	v := &VisionArrayLatchesCommand{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *VisionArrayLatchesCommand:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalVisionRearviewCommand(t *testing.T) {
	v := &VisionRearviewCommand{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONVisionRearviewCommand(t *testing.T) {

	v := &VisionRearviewCommand{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *VisionRearviewCommand:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalTrackerEnable(t *testing.T) {
	v := &TrackerEnable{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONTrackerEnable(t *testing.T) {

	v := &TrackerEnable{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *TrackerEnable:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalDistanceTraveled(t *testing.T) {
	v := &DistanceTraveled{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONDistanceTraveled(t *testing.T) {

	v := &DistanceTraveled{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *DistanceTraveled:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalChargerState(t *testing.T) {
	v := &ChargerState{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONChargerState(t *testing.T) {

	v := &ChargerState{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *ChargerState:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalChargerBmsRequest(t *testing.T) {
	v := &ChargerBmsRequest{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONChargerBmsRequest(t *testing.T) {

	v := &ChargerBmsRequest{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *ChargerBmsRequest:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalChargerCurrentVoltage(t *testing.T) {
	v := &ChargerCurrentVoltage{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONChargerCurrentVoltage(t *testing.T) {

	v := &ChargerCurrentVoltage{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *ChargerCurrentVoltage:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalChargerPower(t *testing.T) {
	v := &ChargerPower{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONChargerPower(t *testing.T) {

	v := &ChargerPower{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *ChargerPower:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalThunderstruckControlMessage(t *testing.T) {
	v := &ThunderstruckControlMessage{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONThunderstruckControlMessage(t *testing.T) {

	v := &ThunderstruckControlMessage{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *ThunderstruckControlMessage:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalVisionStatusFront(t *testing.T) {
	v := &VisionStatusFront{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONVisionStatusFront(t *testing.T) {

	v := &VisionStatusFront{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *VisionStatusFront:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalVisionStatusRear(t *testing.T) {
	v := &VisionStatusRear{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONVisionStatusRear(t *testing.T) {

	v := &VisionStatusRear{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *VisionStatusRear:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalLightsFrontId(t *testing.T) {
	v := &LightsFrontId{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONLightsFrontId(t *testing.T) {

	v := &LightsFrontId{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *LightsFrontId:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalLightsBackId(t *testing.T) {
	v := &LightsBackId{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONLightsBackId(t *testing.T) {

	v := &LightsBackId{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *LightsBackId:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalVisionId(t *testing.T) {
	v := &VisionId{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONVisionId(t *testing.T) {

	v := &VisionId{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *VisionId:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalSteeringPressCount1(t *testing.T) {
	v := &SteeringPressCount1{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONSteeringPressCount1(t *testing.T) {

	v := &SteeringPressCount1{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *SteeringPressCount1:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalSteeringPressCount2(t *testing.T) {
	v := &SteeringPressCount2{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONSteeringPressCount2(t *testing.T) {

	v := &SteeringPressCount2{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *SteeringPressCount2:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalSteeringButtonColors1(t *testing.T) {
	v := &SteeringButtonColors1{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONSteeringButtonColors1(t *testing.T) {

	v := &SteeringButtonColors1{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *SteeringButtonColors1:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalSteeringButtonColors2(t *testing.T) {
	v := &SteeringButtonColors2{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONSteeringButtonColors2(t *testing.T) {

	v := &SteeringButtonColors2{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *SteeringButtonColors2:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalSteeringHorn(t *testing.T) {
	v := &SteeringHorn{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONSteeringHorn(t *testing.T) {

	v := &SteeringHorn{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *SteeringHorn:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalThunderstruckStatusMessage(t *testing.T) {
	v := &ThunderstruckStatusMessage{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONThunderstruckStatusMessage(t *testing.T) {

	v := &ThunderstruckStatusMessage{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *ThunderstruckStatusMessage:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalTrackerData(t *testing.T) {
	v := &TrackerData{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONTrackerData(t *testing.T) {

	v := &TrackerData{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *TrackerData:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalTritiumMotorDrive(t *testing.T) {
	v := &TritiumMotorDrive{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONTritiumMotorDrive(t *testing.T) {

	v := &TritiumMotorDrive{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *TritiumMotorDrive:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalTritiumMotorPower(t *testing.T) {
	v := &TritiumMotorPower{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONTritiumMotorPower(t *testing.T) {

	v := &TritiumMotorPower{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *TritiumMotorPower:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalTritiumReset(t *testing.T) {
	v := &TritiumReset{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONTritiumReset(t *testing.T) {

	v := &TritiumReset{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *TritiumReset:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalBmsAhSet(t *testing.T) {
	v := &BmsAhSet{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONBmsAhSet(t *testing.T) {

	v := &BmsAhSet{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *BmsAhSet:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalBmsWhSet(t *testing.T) {
	v := &BmsWhSet{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONBmsWhSet(t *testing.T) {

	v := &BmsWhSet{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *BmsWhSet:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalBmsKill(t *testing.T) {
	v := &BmsKill{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONBmsKill(t *testing.T) {

	v := &BmsKill{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *BmsKill:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalTelemetryRtcReset(t *testing.T) {
	v := &TelemetryRtcReset{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONTelemetryRtcReset(t *testing.T) {

	v := &TelemetryRtcReset{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *TelemetryRtcReset:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalWsrIdentification(t *testing.T) {
	v := &WsrIdentification{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONWsrIdentification(t *testing.T) {

	v := &WsrIdentification{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *WsrIdentification:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalWsrStatusInformation(t *testing.T) {
	v := &WsrStatusInformation{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONWsrStatusInformation(t *testing.T) {

	v := &WsrStatusInformation{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *WsrStatusInformation:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalWsrBusMeasurement(t *testing.T) {
	v := &WsrBusMeasurement{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONWsrBusMeasurement(t *testing.T) {

	v := &WsrBusMeasurement{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *WsrBusMeasurement:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalWsrVelocity(t *testing.T) {
	v := &WsrVelocity{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONWsrVelocity(t *testing.T) {

	v := &WsrVelocity{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *WsrVelocity:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalWsrPhaseCurrent(t *testing.T) {
	v := &WsrPhaseCurrent{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONWsrPhaseCurrent(t *testing.T) {

	v := &WsrPhaseCurrent{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *WsrPhaseCurrent:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalWsrMotorVoltageVector(t *testing.T) {
	v := &WsrMotorVoltageVector{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONWsrMotorVoltageVector(t *testing.T) {

	v := &WsrMotorVoltageVector{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *WsrMotorVoltageVector:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalWsrMotorCurrentVector(t *testing.T) {
	v := &WsrMotorCurrentVector{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONWsrMotorCurrentVector(t *testing.T) {

	v := &WsrMotorCurrentVector{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *WsrMotorCurrentVector:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalWsrMotorBackemf(t *testing.T) {
	v := &WsrMotorBackemf{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONWsrMotorBackemf(t *testing.T) {

	v := &WsrMotorBackemf{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *WsrMotorBackemf:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalWsr15165VoltageRail(t *testing.T) {
	v := &Wsr15165VoltageRail{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONWsr15165VoltageRail(t *testing.T) {

	v := &Wsr15165VoltageRail{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *Wsr15165VoltageRail:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalWsr2512VoltageRail(t *testing.T) {
	v := &Wsr2512VoltageRail{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONWsr2512VoltageRail(t *testing.T) {

	v := &Wsr2512VoltageRail{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *Wsr2512VoltageRail:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalWsrHeatsinkMotorTemp(t *testing.T) {
	v := &WsrHeatsinkMotorTemp{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONWsrHeatsinkMotorTemp(t *testing.T) {

	v := &WsrHeatsinkMotorTemp{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *WsrHeatsinkMotorTemp:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalWsrDspBoardTemp(t *testing.T) {
	v := &WsrDspBoardTemp{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONWsrDspBoardTemp(t *testing.T) {

	v := &WsrDspBoardTemp{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *WsrDspBoardTemp:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalWsrReserved(t *testing.T) {
	v := &WsrReserved{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONWsrReserved(t *testing.T) {

	v := &WsrReserved{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *WsrReserved:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalWsrOdometerBusAmphoursMeasurement(t *testing.T) {
	v := &WsrOdometerBusAmphoursMeasurement{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONWsrOdometerBusAmphoursMeasurement(t *testing.T) {

	v := &WsrOdometerBusAmphoursMeasurement{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *WsrOdometerBusAmphoursMeasurement:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalWsrSlipSpeedMeasurement(t *testing.T) {
	v := &WsrSlipSpeedMeasurement{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONWsrSlipSpeedMeasurement(t *testing.T) {

	v := &WsrSlipSpeedMeasurement{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *WsrSlipSpeedMeasurement:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalWslIdentification(t *testing.T) {
	v := &WslIdentification{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONWslIdentification(t *testing.T) {

	v := &WslIdentification{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *WslIdentification:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalWslStatusInformation(t *testing.T) {
	v := &WslStatusInformation{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONWslStatusInformation(t *testing.T) {

	v := &WslStatusInformation{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *WslStatusInformation:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalWslBusMeasurement(t *testing.T) {
	v := &WslBusMeasurement{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONWslBusMeasurement(t *testing.T) {

	v := &WslBusMeasurement{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *WslBusMeasurement:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalWslVelocity(t *testing.T) {
	v := &WslVelocity{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONWslVelocity(t *testing.T) {

	v := &WslVelocity{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *WslVelocity:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalWslPhaseCurrent(t *testing.T) {
	v := &WslPhaseCurrent{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONWslPhaseCurrent(t *testing.T) {

	v := &WslPhaseCurrent{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *WslPhaseCurrent:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalWslMotorVoltageVector(t *testing.T) {
	v := &WslMotorVoltageVector{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONWslMotorVoltageVector(t *testing.T) {

	v := &WslMotorVoltageVector{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *WslMotorVoltageVector:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalWslMotorCurrentVector(t *testing.T) {
	v := &WslMotorCurrentVector{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONWslMotorCurrentVector(t *testing.T) {

	v := &WslMotorCurrentVector{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *WslMotorCurrentVector:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalWslMotorBackemf(t *testing.T) {
	v := &WslMotorBackemf{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONWslMotorBackemf(t *testing.T) {

	v := &WslMotorBackemf{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *WslMotorBackemf:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalWsl15165VoltageRail(t *testing.T) {
	v := &Wsl15165VoltageRail{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONWsl15165VoltageRail(t *testing.T) {

	v := &Wsl15165VoltageRail{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *Wsl15165VoltageRail:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalWsl2512VoltageRail(t *testing.T) {
	v := &Wsl2512VoltageRail{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONWsl2512VoltageRail(t *testing.T) {

	v := &Wsl2512VoltageRail{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *Wsl2512VoltageRail:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalWslHeatsinkMotorTemp(t *testing.T) {
	v := &WslHeatsinkMotorTemp{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONWslHeatsinkMotorTemp(t *testing.T) {

	v := &WslHeatsinkMotorTemp{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *WslHeatsinkMotorTemp:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalWslDspBoardTemp(t *testing.T) {
	v := &WslDspBoardTemp{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONWslDspBoardTemp(t *testing.T) {

	v := &WslDspBoardTemp{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *WslDspBoardTemp:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalWslOdometerBusAmphoursMeasurement(t *testing.T) {
	v := &WslOdometerBusAmphoursMeasurement{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONWslOdometerBusAmphoursMeasurement(t *testing.T) {

	v := &WslOdometerBusAmphoursMeasurement{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *WslOdometerBusAmphoursMeasurement:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalWslReserved(t *testing.T) {
	v := &WslReserved{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONWslReserved(t *testing.T) {

	v := &WslReserved{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *WslReserved:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
func TestMarshalUnmarshalWslSlipSpeedMeasurement(t *testing.T) {
	v := &WslSlipSpeedMeasurement{}
	bin, err := v.MarshalPacket()
	if err != nil {
		t.Fatal(err)
	}
	err = v.UnmarshalPacket(bin)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONWslSlipSpeedMeasurement(t *testing.T) {

	v := &WslSlipSpeedMeasurement{}

	rawData, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := v.CANId()
	p, err := FromJson(id, rawData)
	if err != nil {
		t.Fatal(err)
	}

	switch underlying := p.(type) {
	case *WslSlipSpeedMeasurement:
		break
	default:
		t.Fatalf("didn't match type: %T, %v", underlying, underlying)
	}
	

}
