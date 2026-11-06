package utils

import "testing"

type testList []struct {
	name, got, exp string
}

func TestBytes(t *testing.T) {
	tl := testList{
		{"bytes(0)", Bytes(0), "0 B"},
		{"bytes(1)", Bytes(1), "1 B"},
		{"bytes(803)", Bytes(803), "803 B"},
		{"bytes(999)", Bytes(999), "999 B"},

		{"bytes(1024)", Bytes(1024), "1.0 kB"},
		{"bytes(9999)", Bytes(9999), "10 kB"},
		{"bytes(1MB - 1)", Bytes(MByte - Byte), "1000 kB"},

		{"bytes(1MB)", Bytes(1024 * 1024), "1.0 MB"},
		{"bytes(1GB - 1K)", Bytes(GByte - KByte), "1000 MB"},

		{"bytes(1GB)", Bytes(GByte), "1.0 GB"},
		{"bytes(1TB - 1M)", Bytes(TByte - MByte), "1000 GB"},
		{"bytes(10MB)", Bytes(9999 * 1000), "10 MB"},

		{"bytes(1TB)", Bytes(TByte), "1.0 TB"},
		{"bytes(1PB - 1T)", Bytes(PByte - TByte), "999 TB"},

		{"bytes(1PB)", Bytes(PByte), "1.0 PB"},
		{"bytes(1PB - 1T)", Bytes(EByte - PByte), "999 PB"},

		{"bytes(1EB)", Bytes(EByte), "1.0 EB"},
		// Overflows.
		// {"bytes(1EB - 1P)", Bytes((KByte*EByte)-PByte), "1023EB"},

		{"bytes(0)", IBytes(0), "0 B"},
		{"bytes(1)", IBytes(1), "1 B"},
		{"bytes(803)", IBytes(803), "803 B"},
		{"bytes(1023)", IBytes(1023), "1023 B"},

		{"bytes(1024)", IBytes(1024), "1.0 KiB"},
		{"bytes(1MB - 1)", IBytes(MiByte - IByte), "1024 KiB"},

		{"bytes(1MB)", IBytes(1024 * 1024), "1.0 MiB"},
		{"bytes(1GB - 1K)", IBytes(GiByte - KiByte), "1024 MiB"},

		{"bytes(1GB)", IBytes(GiByte), "1.0 GiB"},
		{"bytes(1TB - 1M)", IBytes(TiByte - MiByte), "1024 GiB"},

		{"bytes(1TB)", IBytes(TiByte), "1.0 TiB"},
		{"bytes(1PB - 1T)", IBytes(PiByte - TiByte), "1023 TiB"},

		{"bytes(1PB)", IBytes(PiByte), "1.0 PiB"},
		{"bytes(1PB - 1T)", IBytes(EiByte - PiByte), "1023 PiB"},

		{"bytes(1EiB)", IBytes(EiByte), "1.0 EiB"},
		// Overflows.
		// {"bytes(1EB - 1P)", IBytes((KIByte*EIByte)-PiByte), "1023EB"},

		{"bytes(5.5GiB)", IBytes(5.5 * GiByte), "5.5 GiB"},

		{"bytes(5.5GB)", Bytes(5.5 * GByte), "5.5 GB"},
	}
	for _, test := range tl {
		if test.got != test.exp {
			t.Errorf("On %v, expected '%v', but got '%v'",
				test.name, test.exp, test.got)
		}
	}
}

func TestCalculateModelTransferSize(t *testing.T) {
	// FP32: 1000 params * 4 bytes = 4000
	fp32Size, err := CalculateModelTransferSize(1000, 32, 64)
	if err != nil || fp32Size != 4000 {
		t.Errorf("Expected 4000 for FP32, got %d, err: %v", fp32Size, err)
	}

	// FP16: 1000 params * 2 bytes = 2000
	fp16Size, err := CalculateModelTransferSize(1000, 16, 64)
	if err != nil || fp16Size != 2000 {
		t.Errorf("Expected 2000 for FP16, got %d, err: %v", fp16Size, err)
	}

	// INT8: 1000 params: 1000 bytes + 16 blocks * 4 bytes = 1064
	int8Size, err := CalculateModelTransferSize(1000, 8, 64)
	if err != nil || int8Size != 1064 {
		t.Errorf("Expected 1064 for INT8, got %d, err: %v", int8Size, err)
	}

	// INT4: 1000 params: 500 bytes + 16 blocks * 4 bytes = 564
	int4Size, err := CalculateModelTransferSize(1000, 4, 64)
	if err != nil || int4Size != 564 {
		t.Errorf("Expected 564 for INT4, got %d, err: %v", int4Size, err)
	}

	// Invalid precision
	_, err = CalculateModelTransferSize(1000, 3, 64)
	if err == nil {
		t.Errorf("Expected error for unsupported bit precision")
	}

	// Negative parameters
	_, err = CalculateModelTransferSize(-1, 32, 64)
	if err == nil {
		t.Errorf("Expected error for negative parameter count")
	}
}
