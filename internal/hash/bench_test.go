// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package hash_test

import (
	"hash/fnv"
	"math/rand"
	"testing"

	"go.uber.org/zap/internal/hash"
)

// XXX limit to go 1.7+, or stop using sub-benchmarks...

var rand50 = []string{
	"unbracing",
	"stereotomy",
	"supranervian",
	"moaning",
	"exchangeability",
	"gunyang",
	"sulcation",
	"dariole",
	"archheresy",
	"synchronistically",
	"clips",
	"unsanctioned",
	"Argoan",
	"liparomphalus",
	"layship",
	"Fregatae",
	"microzoology",
	"glaciaria",
	"Frugivora",
	"patterist",
	"Grossulariaceae",
	"lithotint",
	"bargander",
	"opisthographical",
	"cacography",
	"chalkstone",
	"nonsubstantialism",
	"sardonicism",
	"calamiform",
	"lodginghouse",
	"predisposedly",
	"topotypic",
	"broideress",
	"outrange",
	"gingivolabial",
	"monoazo",
	"sparlike",
	"concameration",
	"untoothed",
	"Camorrism",
	"reissuer",
	"soap",
	"palaiotype",
	"countercharm",
	"yellowbird",
	"palterly",
	"writinger",
	"boatfalls",
	"tuglike",
	"underbitten",
}

var rand100 = []string{
	"rooty",
	"malcultivation",
	"degrade",
	"pseudoindependent",
	"stillatory",
	"antiseptize",
	"protoamphibian",
	"antiar",
	"Esther",
	"pseudelminth",
	"superfluitance",
	"teallite",
	"disunity",
	"spirignathous",
	"vergency",
	"myliobatid",
	"inosic",
	"overabstemious",
	"patriarchally",
	"foreimagine",
	"coetaneity",
	"hemimellitene",
	"hyperspatial",
	"aulophyte",
	"electropoion",
	"antitrope",
	"Amarantus",
	"smaltine",
	"lighthead",
	"syntonically",
	"incubous",
	"versation",
	"cirsophthalmia",
	"Ulidian",
	"homoeography",
	"Velella",
	"Hecatean",
	"serfage",
	"Spermaphyta",
	"palatoplasty",
	"electroextraction",
	"aconite",
	"avirulence",
	"initiator",
	"besmear",
	"unrecognizably",
	"euphoniousness",
	"balbuties",
	"pascuage",
	"quebracho",
	"Yakala",
	"auriform",
	"sevenbark",
	"superorganism",
	"telesterion",
	"ensand",
	"nagaika",
	"anisuria",
	"etching",
	"soundingly",
	"grumpish",
	"drillmaster",
	"perfumed",
	"dealkylate",
	"anthracitiferous",
	"predefiance",
	"sulphoxylate",
	"freeness",
	"untucking",
	"misworshiper",
	"Nestorianize",
	"nonegoistical",
	"construe",
	"upstroke",
	"teated",
	"nasolachrymal",
	"Mastodontidae",
	"gallows",
	"radioluminescent",
	"uncourtierlike",
	"phasmatrope",
	"Clunisian",
	"drainage",
	"sootless",
	"brachyfacial",
	"antiheroism",
	"irreligionize",
	"ked",
	"unfact",
	"nonprofessed",
	"milady",
	"conjecture",
	"Arctomys",
	"guapilla",
	"Sassenach",
	"emmetrope",
	"rosewort",
	"raphidiferous",
	"pooh",
	"Tyndallize",
}

var (
	rand100s2 = randCross(rand100, rand100, 1000)
	rand100s3 = randCross(rand100s2, rand100, 10000)
	rand100s4 = randCross(rand100s3, rand100, 100000)
)

var testCorpii = []struct {
	name  string
	words []string
}{
	{
		name: "some stuff I made up",
		words: []string{
			"foo",
			"bar",
			"baz",
			"alpha",
			"bravo",
			"charlie",
			"delta",
		},
	},

	{
		name:  "shuf -n50 /usr/share/dict/words",
		words: rand50,
	},

	{
		name:  "shuf -n100 /usr/share/dict/words",
		words: rand100,
	},

	{
		name:  "(shuf -n100 /usr/share/dict/words)^2",
		words: rand100s2,
	},

	{
		name:  "(shuf -n100 /usr/share/dict/words)^3",
		words: rand100s3,
	},

	{
		name:  "(shuf -n100 /usr/share/dict/words)^4",
		words: rand100s4,
	},
}

func randCross(a, b []string, n int) []string {
	r := make([]string, 0, n)
	k := 0
	for i := 0; i < len(a); i++ {
		for j := 0; j < len(b); j++ {
			c := a[i] + " " + b[j]
			if len(r) < cap(r) {
				r = append(r, c)
			} else {
				ri := 1 + rand.Intn(len(r)-1)
				if ri <= k {
					r[ri] = c
				}
			}
			k++
		}
	}
	return r
}

func runHashBenchmark(b *testing.B, h func(string) uint32) {
	for _, corpus := range testCorpii {
		b.Run(corpus.name, func(b *testing.B) {
			b.ResetTimer()
			for i, j := 0, 0; i < b.N; i++ {
				_ = h(corpus.words[j])
				j++
				if n := len(corpus.words); j >= n {
					j -= n
				}
			}
		})
	}
}

func BenchmarkXSHRR(b *testing.B) {
	runHashBenchmark(b, func(k string) uint32 {
		return hash.XSHRR(k, 4096)
	})
}

func BenchmarkCoreFNV32(b *testing.B) {
	h := fnv.New32()
	runHashBenchmark(b, func(k string) uint32 {
		h.Reset()
		h.Write([]byte(k)) // err == nil always
		return h.Sum32()
	})
}
