
# GO-RAINBOW

Rainbow tables for specific customized name spaces

## Quick start



## Context

Both Rainbow tables and Rule-based brute forced have been used to crack passwords. 

Large **rainbow-tables**  provide significant benefits :
* can test any passwords
* re-usable
* fast
  
But also some drawbacks : 
* extremely large tables,
* success rate depends (mainly) on table size, and is not 100%
* cpu/gpu intensive and lengthy to generate, 
* usually limited to 10 characters max.
* dedicated to one hash function
* no benefit from prior information about the password structure

Rule-based, brute-force tools (John-the-Ripper and the like ...) provide different benefits :
* deterministic, success is garanteed if password match expectations
* no preparation (except for rules & configuration)
* can leverage prior password information (probability, structure, ...)

But has its own drawbacks :
* limited time efficiency beyond the obvious
* not reusable, no off-line preparation possible


## Goals of this package

Current package attempts to find a "best of both worlds" approach to the problem, by adding a flexible, rule-based capability to the traditionnal rainbow-table approach. 

Targeted benefits are :
* table based, with space/time tradeoff inspired by rainbow tables, that can be computed in advance
* rule-driven ( word lists and mangling ) to limit the scope of searches to likely passwords, enabling longer passwords in exchange for shorter unlikely ones,
* weighting mecanisms to prioritize encoding (and search success) of most likely passwords (tbd ?)
* no (practical) password size limits ( but of course, provided rules are restrictive enough ... the name space cannot exceed the total number of hash values !)

## Design principles 

Architecture is based on the rainbow table architecture ( see https://lasec.epfl.ch/pub/lasec/doc/Oech03.pdf )

Each chain has the following structure : 

*H0 -(reduce)-> P1 -(hash)-> H1 -> .../... -> Hn*
* Px are variable length, Hx is fixed length  
* The length of the chain, n, is constant and predefined
* No effort is made to avoid collisions and make a "perfect" table (see why in the above reference paper)
* H0 and Hn define the *Chain*, *Chain*s are stored in a *Rainbow* (table) object.

Before using a newly created *Rainbow* table, you need to compile the various *Reduce* rules that will be applied.
