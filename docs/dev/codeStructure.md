# Code Structure

The webway project consists of a number of Go modules, each of which is
responsible for a specific part of the system and is versioned independently.

The modules are:

- github.com/madsrc/webway
- github.com/madsrc/webway/metadatastore
- github.com/madsrc/webway/agent

## Overall strategy

Each module is structured similar to the approach used for the [WTFDial]
project. This also somewhat vaguely resembles what some people refer to as "Hexagonal
Architecture".

The root of each module contains a package that is referred to as the "domain" and
is mostly devoid of business logic. It is only used to define interfaces and types
used by consumers of the package. One important aspect of the domain package is
that it __must not__ make any imports except for the standard library.

Every other package in the module is considered an implementation of parts of the
domain. These packages are allowed to import external dependencies and expose them
by wrapping them in implementations of the domain interfaces. Implementation packages
are only allowed to communicate with other implementation packages through the domain
interfaces and is thus __not__ allowed to import other implementation packages.

Implementation packages are often, but not always, named after the external dependency
that it wraps.

This use of domain and implementation packages forces the use of dependency injections
and thus makes it easier to test the code.

One common exception to the rule of implementation packages is the `cmd` folder.
folder is neither a domain nor an implementation package, but is used to define
the entry points for an application. These applications will typically import both
the domain and the implementation packages and tie them together. The code stored
in the `cmd` folder is also allowed to import external dependencies if absolutely
necessary, but this should be avoided if possible.

## github.com/madsrc/webway

The core webway module contains the main logic for the system and supposed to
contain code that is to be shared between the other modules.

## github.com/madsrc/webway/metadatastore

The metadatastore module contains logic that is specific to the metadatastore
application. This application is responsible for control plane operations, such as keeping
track of the state of the system.

## github.com/madsrc/webway/agent

The agent module contains code used to build and run webway agent instances. These
instances is responsible for serving the data plane by serving endpoints that speak
the Kafka protocol, and by saving and getting data from object storage.

[WTFDial]: https://github.com/benbjohnson/wtf
