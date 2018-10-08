package meters

const (
	METERTYPE_SBC = "SBC"
)

type SBCProducer struct {
	MeasurementMapping
}

func NewSBCProducer() *SBCProducer {
	/**
	 * Opcodes for Saia Burgess ALE3
	 * http://datenblatt.stark-elektronik.de/saia_burgess/DE_DS_Energymeter-ALE3-with-Modbus.pdf
	 */
	ops := Measurements{
		Import: 28, // double, scaler 100
		Export: 32, // double, scaler 100
		// PartialImport: 30, // double, scaler 100
		// PartialExport: 34, // double, scaler 100

		VoltageL1:       36,
		CurrentL1:       37, // scaler 10
		PowerL1:         38, // scaler 100
		ReactivePowerL1: 39, // scaler 100
		CosphiL1:        40, // scaler 100

		VoltageL2:       41,
		CurrentL2:       42, // scaler 10
		PowerL2:         43, // scaler 100
		ReactivePowerL2: 44, // scaler 100
		CosphiL2:        45, // scaler 100

		VoltageL3:       46,
		CurrentL3:       47, // scaler 10
		PowerL3:         48, // scaler 100
		ReactivePowerL3: 49, // scaler 100
		CosphiL3:        50, // scaler 100

		Power:         51, // scaler 100
		ReactivePower: 52, // scaler 100
	}
	return &SBCProducer{
		MeasurementMapping{ops},
	}
}

func (p *SBCProducer) GetMeterType() string {
	return METERTYPE_SBC
}

func (p *SBCProducer) snip(iec Measurement, readlen uint16) Operation {
	return Operation{
		FuncCode: ReadHoldingReg,
		OpCode:   p.Opcode(iec) - 1, // adjust according to docs
		ReadLen:  readlen,
		IEC61850: iec,
	}
}

// snip16 creates modbus operation for single register
func (p *SBCProducer) snip16(iec Measurement, scaler ...float64) Operation {
	snip := p.snip(iec, 1)

	snip.Transform = RTUUint16ToFloat64 // default conversion
	if len(scaler) > 0 {
		snip.Transform = MakeRTUScaledUint16ToFloat64(scaler[0])
	}

	return snip
}

// snip32 creates modbus operation for double register
func (p *SBCProducer) snip32(iec Measurement, scaler ...float64) Operation {
	snip := p.snip(iec, 2)

	snip.Transform = RTUUint32ToFloat64 // default conversion
	if len(scaler) > 0 {
		snip.Transform = MakeRTUScaledUint32ToFloat64(scaler[0])
	}

	return snip
}

func (p *SBCProducer) Probe() Operation {
	return p.snip16(VoltageL1)
}

func (p *SBCProducer) Produce() (res []Operation) {
	for _, op := range []Measurement{VoltageL1, VoltageL2, VoltageL1} {
		res = append(res, p.snip16(op))
	}

	for _, op := range []Measurement{CurrentL1, CurrentL2, CurrentL1} {
		res = append(res, p.snip16(op, 10))
	}

	for _, op := range []Measurement{
		PowerL1, PowerL2, PowerL1,
		CosphiL1, CosphiL2, CosphiL1,
	} {
		res = append(res, p.snip16(op, 100))
	}

	res = append(res, p.snip32(Import, 100))
	res = append(res, p.snip32(Export, 100))

	return res
}
