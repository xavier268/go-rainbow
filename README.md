
# GO-RAINBOW

Rainbow tables for specific customized name spaces

## Legal warning

#### This package should not be used to crack passwords on machines or systems that you don't own or which you are not legally entitled to test.
No warranty, use at your own risk, etc ...

## Quick start

#### 1. Create an empty Rainbow table

Create an empty rainbow table, using MD5 with chains that will be 1_000 bytes long.

````golang
r := rainbow.New(crypto.MD5, 1_000)
````

Any hash function available in the golang std pachkages can be used, or a custom hash function can easily be implemented, using the hash.Hash interface.

#### 2. Compile the password name space

Specify the name space for the passwords you will be looking for.

For instance, to search in a word list :
````golang
r = r.CompileWordList("words_test.txt").Build()
````

You can combine the various CompileXXX functions. For instance, to search for password that start with 3 symbols exactly from '*' or '+', then a word from the file, between 2 and 5 digits :
````golang
r = r.
    CompileAlphabet("*+",3,3).
    CompileWordList("words_test.txt").
    CompileAlphabet("0123456789",2,5).
    Build()
````
*CompileAlphabet* specifies the alphabet ( utf-8 accepted ... and handled correctly ). Password will be generated with a length included between both provided values(inclusive).

The above example would generate the following passwords :
````
    +*+mouse55
    ***cat5645
    *++cow12345
````

Also notice that when you are finished piping the various *CompileXXX* instructions, you need to call *Build()* to finalize the reduce function.
*Build()* can only be called once, and should be called before starting constructing the chains.

*CompileTransform* provides also a way to apply a transformation to the pasword as it is at this stage, selecting a transformation within a list.

For instance, the table below would include words from the file, 1/3 of them been capitalized. You specify a list of potential transformation, some of them can be nil (no change), and each transformation will be chosen with the same probability.
````golang
// Specify the transformation to be applied
func trf (p []byte) ([]byte) {
		return []byte(strings.ToUpper(string(p)))
    }
    
// Compile it
r.CompileTransform(trf, nil, nil)
````
This is a very powerful and flexible approach - for instance, you can try duplicating passwords, replacing 'a' with '@' or any other transformation, based upon your assumption of what passwords are likely to look like ...

#### 3. Compute the chains.

This is the CPU-time intensive part. 

You can have an idea of the size of your name space size by calling *r.BitLen()* to get the approximate number of bits needed to encode the compiled namespace. The more bits, the more chains will be needed to achieve the same lookup success probability. 

CPU time is driven by the number of chains multiplied by the chain length, storage is driven only by the number of chain, and lookup time is penalized by longer chains. So you can play to stay within your limits ...

````golang
for i := 0; i< nbChainsToCompute; i++ {
    c := r.NewChain)    // create a new chain
    r.AddChain(c)       // add the chain to the table
}
````

#### 4. Save to (load from ) file

To save to file, you just call Save on a io.Writer. 
````golang
err := r.Save(writer)
````
To load saved tables, you must first recreate the same empty table, then call Load. If there are existing chains, the new chains will be dedupliacted and merged.
````golang
r := New(crypto.SHA1, 300_000).
    CompileAlphabet("1234567890",5,30).
    Build().
    Load(reader)
````


#### 5. Use an existing table to lookup a password

````golang
// The hash [] to break
h := []byte{2,5,66,255,12,45,12,11,88,89,12,2,3,4,55,6}

// Lookup h into the Rainbow table r
p , found := r.Lookup(h)
if found {
    fmt.Println("Found password : ",string(p))
} else {
    fmt.Println("Password not found")
}
````

## About this package

Architecture is based on the rainbow table architecture ( see https://lasec.epfl.ch/pub/lasec/doc/Oech03.pdf )


#### 1. Why this package ?

Both Rainbow tables and Rule-based brute forced have historically been used to crack passwords. 

Large **rainbow-tables**  provide significant benefits :
* can test any passwords
* re-usable
* fast
  
But also some drawbacks : 
* extremely large tables,
* success rate depends (mainly) on table size, and is not 100%
* cpu/gpu intensive and lengthy to generate, 
* practically limited to ~10 characters max.
* dedicated to one hash function
* *prior information* about the password structure does not help

Rule-based, brute-force tools (John-the-Ripper and the like ...) provide different benefits :
* deterministic, success is garanteed if password match expectations
* no preparation (except for rules & configuration)
* can leverage prior password information (probability, structure, ...)

But has its own drawbacks :
* limited time efficiency beyond the obvious
* not reusable, no off-line preparation possible


#### 2. Specific goals of this package

Current package attempts to find a "best of both worlds" approach to the problem, by adding a flexible, rule-based capability to the traditionnal "rigid" rainbow-table approach. 

Targeted benefits are :
* table based, with space/time tradeoff inspired by rainbow tables, that can be computed in advance
* rule-driven ( word lists and mangling ) to limit the scope of searches to likely passwords, enabling longer passwords in exchange for shorter unlikely ones,
* weighting mecanisms to prioritize encoding (and search success) of most likely passwords (tbd ?)
* no (practical) password length limits ( but of course, provided rules are restrictive enough ... the name space cannot exceed the total number of hash values !)

## Change log

#### v0.4
    First somewhat stable version, tested.
    Multiple reducers coexists
    Demo runs in 120s on my machine

#### v0.5
    New reducer architecture
    Removed obsolete reducers
    Redesigned to avoid using big.Int during runtime (ok while compiling).
    
#### v0.6
    Changed CompileTransformation API
    Changed entropy generation
    Demo now runs in 52s (same machine).

#### v0.6.1
    Added human readable configuration signature