package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/apex-woot/pdf-stream-engine/interpreter"
)

// This is a sample PDF content stream from a real file.
const sampleStream = `
%% Contents for page 1
%% Original object ID: 111 0
BT
/TT0 1 Tf
9.96 -0 0 9.96 36 769.2 Tm
( )Tj
0.002 Tc -0.002 Tw 0 -74.843 Td
[(P)3.8 (ag)-4 (e )]TJ
0 Tc 0 Tw 2.205 0 Td
(1)Tj
0.506 0 Td
( )Tj
0.004 Tc -0.004 Tw 0.253 0 Td
[(o)-2 (f )]TJ
0.006 Tc -0.006 Tw 1.084 0 Td
(10)Tj
0 Tc 0 Tw 1 0 Td
( )Tj
ET
q
36 778.5 25 -25 re
W n
q
/GS0 gs
24.9999987 0 0 25.0000006 36.0000038 753.5000604 cm
/Im0 Do
Q
Q
BT
/TT1 1 Tf
-0.001 Tc 0.001 Tw 15.96 -0 0 15.96 209.28 740.76 Tm
[(N)1.3 (an)-6.1 (d)2 (i)-2.8 (t)-4.4 (a )]TJ
0 Tc 0 Tw 3.857 0 Td
(G)Tj
-0.001 Tc 0.001 Tw 0.647 0 Td
[(an)-6.1 (gu)-5.5 (l)28.8 (y)-11.4 ( )]TJ
0 Tc 0 Tw 3.301 0 Td
(D)Tj
-0.001 Tc 0.001 Tw 0.707 0 Td
(as)Tj
0 Tc 0 Tw 0.992 0 Td
(,)Tj
0.233 0 Td
( )Tj
-0.003 Tc 0.003 Tw 0.218 0 Td
[(P)-13.1 (h.)-4.1 (D)]TJ
0 Tc 0 Tw 2.158 0 Td
( )Tj
0.003 Tc -0.003 Tw 14.04 -0 0 14.04 189 602.4 Tm
[(P)1.6 (r)19.6 (o)-0.7 (f)4.3 (e)4 (s)9.1 (so)-0.6 (r)11 ( )]TJ
0 Tc 0 Tw 4.65 0 Td
<96>Tj
0.496 0 Td
( )Tj
-0.007 Tc 0.009 Tw 0.222 0 Td
[(Ph)-8.3 (a)-10.5 (r)-7.6 (m)2.7 (a)-10.5 (c)0.5 (e)-5.9 (u)-8.3 (t)-1 (i)-9.2 (ca)-10.4 (l)-6.7 ( S)-6.8 (c)-8.1 (i)-9.2 (e)-5.9 (n)-1.3 (ce)-5.9 (s)]TJ
0 Tc 0 Tw 11.291 0 Td
( )Tj
/TT2 1 Tf
-0.003 Tc -0.003 Tw -11.966 -1.171 Td
[(B)-7.4 (u)2 (t)-6.9 (l)-5.5 (e)-2.2 (r)0.8 ( U)-4.6 (n)-0.6 (i)18.6 (v)13.8 (e)-2.2 (r)0.8 (si)1.5 (t)-6.9 (y)]TJ
0 Tc 0 Tw 7.274 0 Td
( )Tj
0.002 Tc -0.002 Tw 12 -0 0 12 36 553.44 Tm
[(B)3 (io)3 (p)-2 (h)4 (ar)6 (m)4 (ac)3 (eu)4 (tic)3 (s)2 (,)-3 ( d)7 (r)6 (u)4 (g)-4 ( tar)6 (g)6 (eting)6 (,)-3 ( an)10 (tic)3 (anc)3 (er)5.9 ( d)7 (r)6 (u)4 (g)6 ( d)6.9 (el)3 (i)20.1 (v)26 (er)-4 (y)86 (,)-3 ( m)4 (u)4 (l)3 (ti)]TJ
0 Tc 0 Tw (-)Tj
0.002 Tc -0.002 Tw 28.04 0 Td
[(d)7 (r)-4 (u)4 (g)6 ( )-10 (r)16 (es)2 (is)2 (tanc)3 (e)10 (,)-3 ( nano)3 (m)4 (ed)7 (ic)3 (i)10 (ne)]TJ
0 Tc 0 Tw ( )Tj
/TT1 1 Tf
-0.002 Tc 0.002 Tw 11.04 -0 0 11.04 36 516.36 Tm
[(ED)-3.5 (UC)5.8 (A)73.9 (T)-4.3 (IO)-2.6 (N)]TJ
0 Tc 0 Tw 5.446 0 Td
( )Tj
-0.005 Tc 0.005 Tw -5.446 -2.261 Td
[(Ph)-5.8 (.)-1.3 (D)]TJ
0 Tc 0 Tw 2.141 0 Td
( )Tj
1.12 0 Td
( )Tj
-0.003 Tc 0.003 Tw 3.261 0 Td
[(P)2.3 (h)-3.8 (a)-0.7 (r)1.5 (m)-4.3 (a)-0.6 (c)-1.4 (e)-4.6 (u)7 (t)-7.6 (i)-4.2 (c)-1.4 (a)-0.6 (l)]TJ
0 Tc 0 Tw 7.196 0 Td
( )Tj
-0.009 Tc 0.009 Tw 0.217 0 Td
[(S)-17.7 (c)-7.4 (i)-10.3 (en)-2.8 (c)-7.4 (es)]TJ
0 Tc 0 Tw 3.87 0 Td
( )Tj
0.001 Tc -0.001 Tw 0.217 0 Td
[(\()-4 (P)6.3 (ha)3.4 (r)5.4 (ma)3.4 (c)2.7 (e)10.2 (ut)-3.6 (ic)2.7 (s)14.3 (\))]TJ
0 Tc 0 Tw 7.62 0 Td
( )Tj
0.446 0 Td
( )Tj
-0.002 Tc 0.002 Tw 3.261 0 Td
[(Un)-6.7 (i)18.5 (v)39.9 (e)-3.6 (r)2.5 (si)-3.2 (t)-6.5 (y)]TJ
0 Tc 0 Tw 4.728 0 Td
( )Tj
-0.004 Tc 0.004 Tw 0.217 0 Td
(of)Tj
0 Tc 0 Tw 0.891 0 Td
( )Tj
-0.005 Tc 0.005 Tw 0.217 0 Td
[(Pi)-6.2 (t)1.3 (t)-9.6 (s)-2.5 (b)-1 (u)-5.8 (rg)4.1 (h)5 (,)]TJ
0 Tc 0 Tw 5.076 0 Td
( )Tj
-0.092 Tc 0.092 Tw 0.217 0 Td
(PA)Tj
0 Tc 0 Tw 1.174 0 Td
( )Tj
0.522 0 Td
( )Tj
3.261 0 Td
( )Tj
-0.005 Tc 0.005 Tw 0.217 0 Td
(1995)Tj
0 Tc 0 Tw 2.359 0 Td
( )Tj
/TT2 1 Tf
0.001 Tc -0.001 Tw -48.228 -1.522 Td
[(Di)-3.6 (s)7 (s)-3.7 (er)1.9 (t)12.9 (at)2.1 (i)7.2 (o)-0.6 (n)4.7 (:)]TJ
0 Tc 0 Tw 5.478 0 Td
( )Tj
-0.01 Tc 0.01 Tw 0.217 0 Td
[(C)-12.2 (h)-12.4 (a)-11.1 (r)12.7 (ac)-14.7 (t)1.9 (e)-11.1 (r)-9 (i)-3.8 (z)-11.5 (at)-9 (i)-14.6 (o)-11.6 (n)]
TJ
`

func main() {
	fmt.Println("Starting PDF Stream Interpreter...")
	fmt.Println("--- Sample Stream ---")
	fmt.Println(sampleStream)
	fmt.Println("---------------------")

	// Create a reader from our sample string
	reader := strings.NewReader(sampleStream)

	// Create a new interpreter
	interp := interpreter.NewInterpreter()

	// Process the stream
	err := interp.ProcessStream(reader)
	if err != nil {
		log.Fatalf("Error processing stream: %v", err)
	}

	// Get the extracted text
	text := interp.GetText()

	fmt.Println("--- Extracted Text ---")
	fmt.Println(text)
	fmt.Println("----------------------")

	// Expected Output:
	// Hello, this is a literal string.
	// This is on a new line.
	// This is larger and will be popped.
	// This is back to 12pt.
	// Hex String
	// Array of strings for TJ
}
