# pcr0btg
A simple tool to reconstruct the PCR0 value for Boot Guard 1.0

## Build
```
$ go build
```

## Usage
```
$ ./pcr0btg -h
Usage of ./pcr0btg:
  -acm-info uint
        Get this from SACM INFO MSR in coreboot CBnT log. (default 12884901997)
  -acm-status uint
        Get this from BIOSACM_ERRORCODE in coreboot CBnT log. (default 2157019139)
  -image string
        filename of firmware image (default "image.bin")
  -pcr0 string
        Expected PCR0 value.
  -txt-enabled
        Set this, if FW has TXT enabled.
  -xzPath string
        Path to system xz command used for lzma encoding. If unset, an internal lzma implementation is used. (default "xz")
```

### Example
```
$ ./pcr0btg -image image.bin -acm-info 0x30000006d -acm-status 0x80918003 -pcr0 CE601D8E8B04460EE7EFD48FBF9E0E8C7946C75A81CC6E7B3CEF273D3815AD78
0000000000000000000000000000000000000000000000000000000000000003|
0982569F98D8264FF6459AB5F7063A41FE24A67A91293498C0057866A4E84FE9 ->
CE601D8E8B04460EE7EFD48FBF9E0E8C7946C75A81CC6E7B3CEF273D3815AD78
Matches PCR0 value! 🥳
```

The values for `-acm-info` and `-acm-status` can be obtain from coreboot CBnT
log messages, for example:
```
CBnT: SACM INFO MSR (0x13A) raw: 0x000000030000006d <--> acm-info
CBnT:   NEM status:              1
CBnT:   TPM type:                TPM 2.0
CBnT:   TPM success:             1
CBnT:   FACB:                    0
CBnT:   measured boot:           1
CBnT:   verified boot:           1
CBnT:   revoked:                 0
CBnT:   BtG capable:             1
CBnT:   TXT capable:             0
CBnT: BOOTSTATUS (0xA0) raw: 0x8000000000000000
CBnT:   Bios trusted:            0
CBnT:   TXT disabled by policy:  0
CBnT:   Bootguard startup error: 0
CBnT:   TXT ucode or ACM error:  0
CBnT:   TXT measurement type 7:  1
CBnT: ERRORCODE (0x30) raw: 0x00000000
CBnT: BIOSACM_ERRORCODE (0x328) raw: 0x80818003  <--> acm-status
CBnT: BIOSACM_ERRORCODE: TXT ucode or ACM error
CBnT:   AC Module Type:          Boot Guard Error
CBnT:   class:                   0x0
CBnT:   major:                   0x0
CBnT:   External:                0x0
```
