package config

import (
	"fmt"
	"learn_DumboNG/014-DumboNG/crypto"
	"os"
	"testing"
)

func TestKeyFiles(t *testing.T) {
	dir := t.TempDir()
	GenerateKeyFiles(4, dir)

	filename := fmt.Sprintf("%s/.node-key-0.json", dir)
	pub, pri, err := GenKeysFromFile(filename)
	if err != nil {
		t.Fatal(err)
	}

	srvc := crypto.NewSigService(pri, crypto.SecretShareKey{})
	d := crypto.NewHasher().Sum256([]byte("dcz"))
	sig, _ := srvc.RequestSignature(d)
	if !sig.Verify(pub, d) {
		t.Fatalf("signature verification failed")
	}
}

func TestThresholdKeyFiles(t *testing.T) {
	dir := t.TempDir()
	GenerateTsKeyFiles(4, 3, dir)

	var shareKeys []crypto.SecretShareKey
	for i := 0; i < 4; i++ {
		filename := fmt.Sprintf("%s/.node-ts-key-%d.json", dir, i)
		shareKey, err := GenTsKeyFromFile(filename)
		if err != nil {
			t.Fatal(err)
		}
		shareKeys = append(shareKeys, shareKey)
	}

	var (
		cnt        = 0
		shareSigch = make(chan crypto.SignatureShare, 4)
	)

	digest := crypto.NewHasher().Sum256([]byte("dczhahahah"))
	for i := 0; i < 4; i++ {
		shareKey := shareKeys[i]
		byt, err := crypto.EncodeTSPartialKey(shareKey.PriShare)
		if err != nil {
			t.Fatal(err)
		}

		share, err := crypto.DecodeTSPartialKey(byt)
		if err != nil {
			t.Fatal(err)
		}
		if share.String() != shareKey.PriShare.String() {
			t.Fatal("encode/decode error")
		}

		srvc := crypto.NewSigService(crypto.PrivateKey{}, shareKey)
		sigShare, err := srvc.RequestTsSugnature(digest)
		if err != nil {
			t.Fatal(err)
		}
		shareSigch <- sigShare
	}

	var sigs []crypto.SignatureShare
	for sig := range shareSigch {
		sigs = append(sigs, sig)
		cnt++
		if cnt == 3 {
			break
		}
	}
	combineSig, err := crypto.CombineIntactTSPartial(sigs, shareKeys[0], digest)
	if err != nil {
		t.Fatal(err)
	}
	if err := crypto.VerifyTs(shareKeys[0], digest, combineSig); err != nil {
		t.Fatal(err)
	}
}

func TestSampleFiles(t *testing.T) {
	dir := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatal(err)
		}
	}()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	GenerateSampleParameters()
	GenerateSmapleCommittee()

	poolP, coreP, err := GenParamatersFromFile("./.parameters.json")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(poolP, coreP)

	committee, err := GenCommitteeFromFile("./.committee.json")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(committee)
}
