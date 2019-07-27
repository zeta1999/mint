package mint

import (
	"crypto/x509"
	"testing"
)

func TestCTLSRecordLayer(t *testing.T) {
	configServer := &Config{
		RequireClientAuth: true,
		Certificates:      certificates,
		RecordLayer:       CTLSRecordLayerFactory{IsServer: true},
	}
	configClient := &Config{
		ServerName:         serverName,
		Certificates:       clientCertificates,
		InsecureSkipVerify: true,
		RecordLayer:        CTLSRecordLayerFactory{IsServer: false},
	}

	cConn, sConn := pipe()
	client := Client(cConn, configClient)
	server := Server(sConn, configServer)

	var clientAlert, serverAlert Alert
	done := make(chan bool)
	go func(t *testing.T) {
		serverAlert = server.Handshake()
		assertEquals(t, serverAlert, AlertNoAlert)
		done <- true
	}(t)

	clientAlert = client.Handshake()
	assertEquals(t, clientAlert, AlertNoAlert)

	<-done

	checkConsistency(t, client, server)
	assertTrue(t, client.state.Params.UsingClientAuth, "Session did not negotiate client auth")
}

func TestCTLSRPK(t *testing.T) {
	suite := TLS_AES_128_GCM_SHA256
	group := X25519
	scheme := ECDSA_P256_SHA256
	virtualFinished := false
	shortRandom := true
	randomSize := 16

	allCertificates := map[string]*Certificate{
		"a": {
			Chain:      []*x509.Certificate{serverCert},
			PrivateKey: serverKey,
		},
		"b": {
			Chain:      []*x509.Certificate{clientCert},
			PrivateKey: clientKey,
		},
	}

	compression := &RPKCompression{
		SupportedVersion: tls13Version,
		ServerName:       serverName,
		CipherSuite:      suite,
		SignatureScheme:  scheme,
		SupportedGroup:   group,
		Certificates:     allCertificates,
		RandomSize:       randomSize,
		VirtualFinished:  virtualFinished,
	}

	configServer := &Config{
		RequireClientAuth: true,
		Certificates:      certificates,
		CipherSuites:      []CipherSuite{suite},
		SignatureSchemes:  []SignatureScheme{scheme},
		Groups:            []NamedGroup{group},
		ShortRandom:       shortRandom,
		RandomSize:        randomSize,
		VirtualFinished:   virtualFinished,
		RecordLayer: CTLSRecordLayerFactory{
			IsServer:    true,
			Compression: compression,
		},
	}
	configClient := &Config{
		ServerName:         serverName,
		Certificates:       clientCertificates,
		InsecureSkipVerify: true,
		CipherSuites:       []CipherSuite{suite},
		SignatureSchemes:   []SignatureScheme{scheme},
		Groups:             []NamedGroup{group},
		ShortRandom:        shortRandom,
		RandomSize:         randomSize,
		VirtualFinished:    virtualFinished,
		RecordLayer: CTLSRecordLayerFactory{
			IsServer:    false,
			Compression: compression,
		},
	}

	cConn, sConn := pipe()
	client := Client(cConn, configClient)
	server := Server(sConn, configServer)

	var clientAlert, serverAlert Alert
	done := make(chan bool)
	go func(t *testing.T) {
		serverAlert = server.Handshake()
		assertEquals(t, serverAlert, AlertNoAlert)
		done <- true
	}(t)

	clientAlert = client.Handshake()
	assertEquals(t, clientAlert, AlertNoAlert)

	<-done

	checkConsistency(t, client, server)
	assertTrue(t, client.state.Params.UsingClientAuth, "Session did not negotiate client auth")
}

func TestCTLSPSK(t *testing.T) {
	suite := TLS_AES_128_GCM_SHA256
	group := X25519
	scheme := ECDSA_P256_SHA256
	pskMode := PSKModeDHEKE
	virtualFinished := false
	shortRandom := true
	randomSize := 16

	psk = PreSharedKey{
		CipherSuite:  TLS_AES_128_GCM_SHA256,
		IsResumption: false,
		Identity:     []byte{0, 1, 2, 3},
		Key:          []byte{4, 5, 6, 7},
	}

	psks = &PSKMapCache{
		serverName: psk,
		"00010203": psk,
	}

	compression := &PSKCompression{
		SupportedVersion: tls13Version,
		ServerName:       serverName,
		CipherSuite:      suite,
		SupportedGroup:   group,
		SignatureScheme:  scheme,
		PSKMode:          pskMode,
		RandomSize:       randomSize,
		VirtualFinished:  virtualFinished,
	}

	configClient := &Config{
		ServerName:       serverName,
		CipherSuites:     []CipherSuite{TLS_AES_128_GCM_SHA256},
		PSKs:             psks,
		Groups:           []NamedGroup{group},
		SignatureSchemes: []SignatureScheme{scheme},
		PSKModes:         []PSKKeyExchangeMode{pskMode},
		ShortRandom:      shortRandom,
		RandomSize:       randomSize,
		VirtualFinished:  virtualFinished,
		RecordLayer: CTLSRecordLayerFactory{
			IsServer:    false,
			Compression: compression,
		},
	}
	configServer := &Config{
		ServerName:       serverName,
		CipherSuites:     []CipherSuite{TLS_AES_128_GCM_SHA256},
		PSKs:             psks,
		Groups:           []NamedGroup{group},
		SignatureSchemes: []SignatureScheme{scheme},
		PSKModes:         []PSKKeyExchangeMode{pskMode},
		ShortRandom:      shortRandom,
		RandomSize:       randomSize,
		VirtualFinished:  virtualFinished,
		RecordLayer: CTLSRecordLayerFactory{
			IsServer:    true,
			Compression: compression,
		},
	}

	cConn, sConn := pipe()
	client := Client(cConn, configClient)
	server := Server(sConn, configServer)

	var clientAlert, serverAlert Alert
	done := make(chan bool)
	go func(t *testing.T) {
		serverAlert = server.Handshake()
		t.Logf("server alert: %v", serverAlert)
		assertEquals(t, serverAlert, AlertNoAlert)
		done <- true
	}(t)

	clientAlert = client.Handshake()
	assertEquals(t, clientAlert, AlertNoAlert)

	<-done

	checkConsistency(t, client, server)
}
