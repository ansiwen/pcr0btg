package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/linuxboot/fiano/pkg/intel/metadata/bg/bgbootpolicy"
	"github.com/linuxboot/fiano/pkg/intel/metadata/bg/bgkey"
	"github.com/linuxboot/fiano/pkg/intel/metadata/fit"
)

var (
	ACMInfo       = flag.Uint64("acm-info", 0x30000006D, "Get this from SACM INFO MSR in coreboot CBnT log.")
	ACMStatus     = flag.Uint64("acm-status", 0x80918003, "Get this from BIOSACM_ERRORCODE in coreboot CBnT log.")
	TXTEnabled    = flag.Bool("txt-enabled", false, "Set this, if FW has TXT enabled.")
	imageFileName = flag.String("image", "image.bin", "filename of firmware image")
	ExpectedPCR0  = flag.String("pcr0", "", "Expected PCR0 value.")
)

func bit(n uint64, i int) byte {
	return byte((n >> i) & 1)
}

func getRSTR() byte {
	var b byte
	b |= bit(*ACMInfo, 4)         // FACB
	b |= bit(*ACMStatus, 21) << 1 // Minor Error Code
	b |= bit(*ACMStatus, 22) << 2 // |
	b |= bit(*ACMStatus, 23) << 3 // |
	b |= bit(*ACMStatus, 24) << 4 // |
	b |= 1 << 6
	return b
}

func getTYPE() byte {
	var b byte
	b |= bit(*ACMInfo, 5)         // Measured
	b |= bit(*ACMInfo, 6) << 1    // Verified
	b |= bit(*ACMStatus, 20) << 2 // Minor Error Code
	if *TXTEnabled {
		b |= 1 << 3
	}
	return b
}

func extend(l []byte, r []byte) []byte {
	h := sha256.New()
	h.Write(l)
	h.Write(r)
	res := h.Sum(nil)
	fmt.Printf("%X|\n%X ->\n%X\n", l, r, res)
	return res
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	flag.Parse()

	image, err := os.ReadFile(*imageFileName)
	check(err)

	fitEntries, err := fit.GetEntries(image)
	check(err)

	var (
		// acmEntry *fit.EntrySACM
		acmData *fit.EntrySACMData
		// kmEntry  *fit.EntryKeyManifestRecord
		kmData *bgkey.Manifest
		// bpmEntry *fit.EntryBootPolicyManifestRecord
		bpmData *bgbootpolicy.Manifest
	)
	for _, fitEntry := range fitEntries {
		switch fitEntry := fitEntry.(type) {
		case *fit.EntrySACM:
			// acmEntry = fitEntry
			acmData, err = fitEntry.ParseData()
			check(err)
		case *fit.EntryKeyManifestRecord:
			// kmEntry = fitEntry
			kmData, _, err = fitEntry.ParseData()
			check(err)
			if kmData == nil {
				log.Fatal("No BtG 1.0 Key Manifest")
			}
		case *fit.EntryBootPolicyManifestRecord:
			// bpmEntry = fitEntry
			bpmData, _, err = fitEntry.ParseData()
			check(err)
		}
	}

	if acmData == nil {
		log.Fatal("No ACM entry in FIT")
	}

	if kmData == nil {
		log.Fatal("No KM entry in FIT")
	}

	if bpmData == nil {
		log.Fatal("No BPM entry in FIT")
	}

	ACMSVN := binary.LittleEndian.AppendUint16(nil, uint16(acmData.GetTXTSVN()))
	ACMSig := acmData.GetRSASig()
	KMSig := kmData.KeyAndSignature.Signature.Data
	BPMSig := bpmData.PMSE.KeySignature.Signature.Data
	IBBHash := bpmData.SE[0].Digest.HashBuffer

	h := sha256.New()
	h.Write([]byte{getRSTR(), getTYPE()})
	h.Write(ACMSVN)
	h.Write(ACMSig)
	h.Write(KMSig)
	h.Write(BPMSig)
	h.Write(IBBHash)
	sum := h.Sum(nil)

	pcr := make([]byte, 32)
	if bpmData.SE[0].Flags.Locality3Startup() {
		pcr[31] = 3
	} else {
		log.Println(`
	Warning: PCR0 is initialized with Locality 0. Measurements can be spoofed,
	if initialization can be skipped, for example if the Chipset is not fused
	to a profile that enforces a measurement.
		`)
	}

	pcr = extend(pcr, sum)
	if *ExpectedPCR0 != "" {
		expected, err := hex.DecodeString(*ExpectedPCR0)
		check(err)
		if bytes.Equal(pcr, expected) {
			fmt.Println("Matches PCR0 value! 🥳")
		}
	}
}
