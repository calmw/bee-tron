// Copyright 2018 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package encryption_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/calmw/bee-tron/pkg/encryption"
	"github.com/calmw/bee-tron/pkg/swarm"
	"github.com/calmw/bee-tron/pkg/util/testutil"
	"golang.org/x/crypto/sha3"
)

var expectedTransformedHex = "352187af3a843decc63ceca6cb01ea39dbcf77caf0a8f705f5c30d557044ceec9392b94a79376f1e5c10cd0c0f2a98e5353bf22b3ea4fdac6677ee553dec192e3db64e179d0474e96088fb4abd2babd67de123fb398bdf84d818f7bda2c1ab60b3ea0e0569ae54aa969658eb4844e6960d2ff44d7c087ee3aaffa1c0ee5df7e50b615f7ad90190f022934ad5300c7d1809bfe71a11cc04cece5274eb97a5f20350630522c1dbb7cebaf4f97f84e03f5cfd88f2b48880b25d12f4d5e75c150f704ef6b46c72e07db2b705ac3644569dccd22fd8f964f6ef787fda63c46759af334e6f665f70eac775a7017acea49f3c7696151cb1b9434fa4ac27fb803921ffb5ec58dafa168098d7d5b97e384be3384cf5bc235c3d887fef89fe76c0065f9b8d6ad837b442340d9e797b46ef5709ea3358bc415df11e4830de986ef0f1c418ffdcc80e9a3cda9bea0ab5676c0d4240465c43ba527e3b4ea50b4f6255b510e5d25774a75449b0bd71e56c537ade4fcf0f4d63c99ae1dbb5a844971e2c19941b8facfcfc8ee3056e7cb3c7114c5357e845b52f7103cb6e00d2308c37b12baa5b769e1cc7b00fc06f2d16e70cc27a82cb9c1a4e40cb0d43907f73df2c9db44f1b51a6b0bc6d09f77ac3be14041fae3f9df2da42df43ae110904f9ecee278030185254d7c6e918a5512024d047f77a992088cb3190a6587aa54d0c7231c1cd2e455e0d4c07f74bece68e29cd8ba0190c0bcfb26d24634af5d91a81ef5d4dd3d614836ce942ddbf7bb1399317f4c03faa675f325f18324bf9433844bfe5c4cc04130c8d5c329562b7cd66e72f7355de8f5375a72202971613c32bd7f3fcdcd51080758cd1d0a46dbe8f0374381dbc359f5864250c63dde8131cbd7c98ae2b0147d6ea4bf65d1443d511b18e6d608bbb46ac036353b4c51df306a10a6f6939c38629a5c18aaf89cac04bd3ad5156e6b92011c88341cb08551bab0a89e6a46538f5af33b86121dba17e3a434c273f385cd2e8cb90bdd32747d8425d929ccbd9b0815c73325988855549a8489dfd047daf777aaa3099e54cf997175a5d9e1edfe363e3b68c70e02f6bf4fcde6a0f3f7d0e7e98bde1a72ae8b6cd27b32990680cc4a04fc467f41c5adcaddabfc71928a3f6872c360c1d765260690dd28b269864c8e380d9c92ef6b89b0094c8f9bb22608b4156381b19b920e9583c9616ce5693b4d2a6c689f02e6a91584a8e501e107403d2689dd0045269dd9946c0e969fb656a3b39d84a798831f5f9290f163eb2f97d3ae25071324e95e2256d9c1e56eb83c26397855323edc202d56ad05894333b7f0ed3c1e4734782eb8bd5477242fd80d7a89b12866f85cfae476322f032465d6b1253993033fccd4723530630ab97a1566460af9c90c9da843c229406e65f3fa578bd6bf04dee9b6153807ddadb8ceefc5c601a8ab26023c67b1ab1e8e0f29ce94c78c308005a781853e7a2e0e51738939a657c987b5e611f32f47b5ff461c52e63e0ea390515a8e1f5393dae54ea526934b5f310b76e3fa050e40718cb4c8a20e58946d6ee1879f08c52764422fe542b3240e75eccb7aa75b1f8a651e37a3bc56b0932cdae0e985948468db1f98eb4b77b82081ea25d8a762db00f7898864984bd80e2f3f35f236bf57291dec28f550769943bcfb6f884b7687589b673642ef7fe5d7d5a87d3eca5017f83ccb9a3310520474479464cb3f433440e7e2f1e28c0aef700a45848573409e7ab66e0cfd4fe5d2147ace81bc65fd8891f6245cd69246bbf5c27830e5ab882dd1d02aba34ff6ca9af88df00fd602892f02fedbdc65dedec203faf3f8ff4a97314e0ddb58b9ab756a61a562597f4088b445fcc3b28a708ca7b1485dcd791b779fbf2b3ef1ec5c6205f595fbe45a02105034147e5a146089c200a49dae33ae051a08ea5f974a21540aaeffa7f9d9e3d35478016fb27b871036eb27217a5b834b461f535752fb5f1c8dded3ae14ce3a2ef6639e2fe41939e3509e46e347a95d50b2080f1ba42c804b290ddc912c952d1cec3f2661369f738feacc0dbf1ea27429c644e45f9e26f30c341acd34c7519b2a1663e334621691e810767e9918c2c547b2e23cce915f97d26aac8d0d2fcd3edb7986ad4e2b8a852edebad534cb6c0e9f0797d3563e5409d7e068e48356c67ce519246cd9c560e881453df97cbba562018811e6cf8c327f399d1d1253ab47a19f4a0ccc7c6d86a9603e0551da310ea595d71305c4aad96819120a92cdbaf1f77ec8df9cc7c838c0d4de1e8692dd81da38268d1d71324bcffdafbe5122e4b81828e021e936d83ae8021eac592aa52cd296b5ce392c7173d622f8e07d18f59bb1b08ba15211af6703463b09b593af3c37735296816d9f2e7a369354a5374ea3955e14ca8ac56d5bfe4aef7a21bd825d6ae85530bee5d2aaaa4914981b3dfdb2e92ec2a27c83d74b59e84ff5c056f7d8945745f2efc3dcf28f288c6cd8383700fb2312f7001f24dd40015e436ae23e052fe9070ea9535b9c989898a9bda3d5382cf10e432fae6ccf0c825b3e6436edd3a9f8846e5606f8563931b5f29ba407c5236e5730225dda211a8504ec1817bc935e1fd9a532b648c502df302ed2063aed008fd5676131ac9e95998e9447b02bd29d77e38fcfd2959f2de929b31970335eb2a74348cc6918bc35b9bf749eab0fe304c946cd9e1ca284e6853c42646e60b6b39e0d3fb3c260abfc5c1b4ca3c3770f344118ca7c7f5c1ad1f123f8f369cd60afc3cdb3e9e81968c5c9fa7c8b014ffe0508dd4f0a2a976d5d1ca8fc9ad7a237d92cfe7b41413d934d6e142824b252699397e48e4bac4e91ebc10602720684bd0863773c548f9a2f9724245e47b129ecf65afd7252aac48c8a8d6fd3d888af592a01fb02dc71ed7538a700d3d16243e4621e0fcf9f8ed2b4e11c9fa9a95338bb1dac74a7d9bc4eb8cbf900b634a2a56469c00f5994e4f0934bdb947640e6d67e47d0b621aacd632bfd3c800bd7d93bd329f494a90e06ed51535831bd6e07ac1b4b11434ef3918fa9511813a002913f33f836454798b8d1787fea9a4c4743ba091ed192ed92f4d33e43a226bf9503e1a83a16dd340b3cbbf38af6db0d99201da8de529b4225f3d2fa2aad6621afc6c79ef3537720591edfc681ae6d00ede53ed724fc71b23b90d2e9b7158aaee98d626a4fe029107df2cb5f90147e07ebe423b1519d848af18af365c71bfd0665db46be493bbe99b79a188de0cf3594aef2299f0324075bdce9eb0b87bc29d62401ba4fd6ae48b1ba33261b5b845279becf38ee03e3dc5c45303321c5fac96fd02a3ad8c9e3b02127b320501333c9e6360440d1ad5e64a6239501502dde1a49c9abe33b66098458eee3d611bb06ffcd234a1b9aef4af5021cd61f0de6789f822ee116b5078aae8c129e8391d8987500d322b58edd1595dc570b57341f2df221b94a96ab7fbcf32a8ca9684196455694024623d7ed49f7d66e8dd453c0bae50e0d8b34377b22d0ece059e2c385dfc70b9089fcd27577c51f4d870b5738ee2b68c361a67809c105c7848b68860a829f29930857a9f9d40b14fd2384ac43bafdf43c0661103794c4bd07d1cfdd4681b6aeaefad53d4c1473359bcc5a83b09189352e5bb9a7498dd0effb89c35aad26954551f8b0621374b449bf515630bd3974dca982279733470fdd059aa9c3df403d8f22b38c4709c82d8f12b888e22990350490e16179caf406293cc9e65f116bafcbe96af132f679877061107a2f690a82a8cb46eea57a90abd23798c5937c6fe6b17be3f9bfa01ce117d2c268181b9095bf49f395fea07ca03838de0588c5e2db633e836d64488c1421e653ea52d810d096048c092d0da6e02fa6613890219f51a76148c8588c2487b171a28f17b7a299204874af0131725d793481333be5f08e86ca837a226850b0c1060891603bfecf9e55cddd22c0dbb28d495342d9cc3de8409f72e52a0115141cffe755c74f061c1a770428ccb0ae59536ee6fc074fbfc6cacb51a549d327527e20f8407477e60355863f1153f9ce95641198663c968874e7fdb29407bd771d94fdda8180cbb0358f5874738db705924b8cbe0cd5e1484aeb64542fe8f38667b7c34baf818c63b1e18440e9fba575254d063fd49f24ef26432f4eb323f3836972dca87473e3e9bb26dc3be236c3aae6bc8a6da567442309da0e8450e242fc9db836e2964f2c76a3b80a2c677979882dda7d7ebf62c93664018bcf4ec431fe6b403d49b3b36618b9c07c2d0d4569cb8d52223903debc72ec113955b206c34f1ae5300990ccfc0180f47d91afdb542b6312d12aeff7e19c645dc0b9fe6e3288e9539f6d5870f99882df187bfa6d24d179dfd1dac22212c8b5339f7171a3efc15b760fed8f68538bc5cbd845c2d1ab41f3a6c692820653eaef7930c02fbe6061d93805d73decdbb945572a7c44ed0241982a6e4d2d730898f82b3d9877cb7bca41cc6dcee67aa0c3d6db76f0b0a708ace0031113e48429de5d886c10e9200f68f32263a2fbf44a5992c2459fda7b8796ba796e3a0804fc25992ed2c9a5fe0580a6b809200ecde6caa0364b58be11564dcb9a616766dd7906db5636ee708b0204f38d309466d8d4a162965dd727e29f5a6c133e9b4ed5bafe803e479f9b2a7640c942c4a40b14ac7dc9828546052761a070f6404008f1ec3605836339c3da95a00b4fd81b2cabf88b51d2087d5b83e8c5b69bf96d8c72cbd278dad3bbb42b404b436f84ad688a22948adf60a81090f1e904291503c16e9f54b05fc76c881a5f95f0e732949e95d3f1bae2d3652a14fe0dda2d68879604657171856ef72637def2a96ac47d7b3fe86eb3198f5e0e626f06be86232305f2ae79ffcd2725e48208f9d8d63523f81915acc957563ab627cd6bc68c2a37d59fb0ed77a90aa9d085d6914a8ebada22a2c2d471b5163aeddd799d90fbb10ed6851ace2c4af504b7d572686700a59d6db46d5e42bb83f8e0c0ffe1dfa6582cc0b34c921ff6e85e83188d24906d5c08bb90069639e713051b3102b53e6f703e8210017878add5df68e6f2b108de279c5490e9eef5590185c4a1c744d4e00d244e1245a8805bd30407b1bc488db44870ccfd75a8af104df78efa2fb7ba31f048a263efdb3b63271fff4922bece9a71187108f65744a24f4947dc556b7440cb4fa45d296bb7f724588d1f245125b21ea063500029bd49650237f53899daf1312809552c81c5827341263cc807a29fe84746170cdfa1ff3838399a5645319bcaff674bb70efccdd88b3d3bb2f2d98111413585dc5d5bd5168f43b3f55e58972a5b2b9b3733febf02f931bd436648cb617c3794841aab961fe41277ab07812e1d3bc4ff6f4350a3e615bfba08c3b9480ef57904d3a16f7e916345202e3f93d11f7a7305170cb8c4eb9ac88ace8bbd1f377bdd5855d3162d6723d4435e84ce529b8f276a8927915ac759a0d04e5ca4a9d3da6291f0333b475df527e99fe38f7a4082662e8125936640c26dd1d17cf284ce6e2b17777a05aa0574f7793a6a062cc6f7263f7ab126b4528a17becfdec49ac0f7d8705aa1704af97fb861faa8a466161b2b5c08a5bacc79fe8500b913d65c8d3c52d1fd52d2ab2c9f52196e712455619c1cd3e0f391b274487944240e2ed8858dd0823c801094310024ae3fe4dd1cf5a2b6487b42cc5937bbafb193ee331d87e378258963d49b9da90899bbb4b88e79f78e866b0213f4719f67da7bcc2fce073c01e87c62ea3cdbcd589cfc41281f2f4c757c742d6d1e"

var hashFunc = sha3.NewLegacyKeccak256
var testKey encryption.Key

// nolint:gochecknoinits
func init() {
	testKey = swarm.MustParseHexAddress("8abf1502f557f15026716030fb6384792583daf39608a3cd02ff2f47e9bc6e49").Bytes()
}

func TestEncryptDataLongerThanPadding(t *testing.T) {
	t.Parallel()

	enc := encryption.New(testKey, 4095, uint32(0), hashFunc)

	data := make([]byte, 4096)

	expectedError := "data length longer than padding, data length 4096 padding 4095"

	_, err := enc.Encrypt(data)
	if err == nil || err.Error() != expectedError {
		t.Fatalf("Expected error \"%v\" got \"%v\"", expectedError, err)
	}
}

func TestEncryptDataZeroPadding(t *testing.T) {
	t.Parallel()

	enc := encryption.New(testKey, 0, uint32(0), hashFunc)

	data := make([]byte, 2048)

	encrypted, err := enc.Encrypt(data)
	if err != nil {
		t.Fatalf("Expected no error got %v", err)
	}
	if len(encrypted) != 2048 {
		t.Fatalf("Encrypted data length expected \"%v\" got %v", 2048, len(encrypted))
	}
}

func TestEncryptDataLengthEqualsPadding(t *testing.T) {
	t.Parallel()

	enc := encryption.New(testKey, 4096, uint32(0), hashFunc)

	data := make([]byte, 4096)

	encrypted, err := enc.Encrypt(data)
	if err != nil {
		t.Fatalf("Expected no error got %v", err)
	}
	encryptedHex := hex.EncodeToString(encrypted)
	expectedTransformed, _ := hex.DecodeString(expectedTransformedHex)

	if !bytes.Equal(encrypted, expectedTransformed) {
		t.Fatalf("Expected %v got %v", expectedTransformedHex, encryptedHex)
	}
}

func TestEncryptDataLengthSmallerThanPadding(t *testing.T) {
	t.Parallel()

	enc := encryption.New(testKey, 4096, uint32(0), hashFunc)

	data := make([]byte, 4080)

	encrypted, err := enc.Encrypt(data)
	if err != nil {
		t.Fatalf("Expected no error got %v", err)
	}
	if len(encrypted) != 4096 {
		t.Fatalf("Encrypted data length expected %v got %v", 4096, len(encrypted))
	}
}

func TestEncryptDataCounterNonZero(t *testing.T) {
	t.Parallel()

	// TODO
}

func TestDecryptDataLengthNotEqualsPadding(t *testing.T) {
	t.Parallel()

	enc := encryption.New(testKey, 4096, uint32(0), hashFunc)

	data := make([]byte, 4097)

	expectedError := "data length different than padding, data length 4097 padding 4096"

	_, err := enc.Decrypt(data)
	if err == nil || err.Error() != expectedError {
		t.Fatalf("Expected error \"%v\" got \"%v\"", expectedError, err)
	}
}

func TestEncryptDecryptIsIdentity(t *testing.T) {
	t.Parallel()

	testEncryptDecryptIsIdentity(t, 0, 2048, 2048, 32)
	testEncryptDecryptIsIdentity(t, 0, 4096, 4096, 32)
	testEncryptDecryptIsIdentity(t, 0, 4096, 1000, 32)
	testEncryptDecryptIsIdentity(t, 10, 32, 32, 32)
}

func testEncryptDecryptIsIdentity(t *testing.T, initCtr uint32, padding, dataLength, keyLength int) {
	t.Helper()

	key := encryption.GenerateRandomKey(keyLength)
	enc := encryption.New(key, padding, initCtr, hashFunc)

	data := testutil.RandBytesWithSeed(t, dataLength, 1)

	encrypted, err := enc.Encrypt(data)
	if err != nil {
		t.Fatalf("Expected no error got %v", err)
	}

	enc.Reset()
	decrypted, err := enc.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Expected no error got %v", err)
	}
	if len(decrypted) != padding {
		t.Fatalf("Expected decrypted data length %v got %v", padding, len(decrypted))
	}

	// we have to remove the extra bytes which were randomly added to fill until padding
	if len(data) < padding {
		decrypted = decrypted[:len(data)]
	}

	if !bytes.Equal(data, decrypted) {
		t.Fatalf("Expected decrypted %v got %v", hex.EncodeToString(data), hex.EncodeToString(decrypted))
	}
}

// TestEncryptSectioned tests that the cipherText is the same regardless of size of data input buffer
func TestEncryptSectioned(t *testing.T) {
	t.Parallel()

	data := testutil.RandBytes(t, 4096)
	key := testutil.RandBytes(t, encryption.KeyLength)

	enc := encryption.New(key, 0, uint32(42), sha3.NewLegacyKeccak256)
	whole, err := enc.Encrypt(data)
	if err != nil {
		t.Fatal(err)
	}

	enc.Reset()
	for i := 0; i < 4096; i += encryption.KeyLength {
		cipher, err := enc.Encrypt(data[i : i+encryption.KeyLength])
		if err != nil {
			t.Fatal(err)
		}
		wholeSection := whole[i : i+encryption.KeyLength]
		if !bytes.Equal(cipher, wholeSection) {
			t.Fatalf("index %d, expected %x, got %x", i/encryption.KeyLength, wholeSection, cipher)
		}
	}
}
