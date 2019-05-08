# pigo-openfaas
OpenFaaS function for Pigo face detector experiments.

### Usage
To run the function locally you have to make sure OpenFaaS is up and running. Read the official documentation for more help. https://docs.openfaas.com/

Clone the repository:
```bash
$ git clone https://github.com/dfquaresma/pigo-openfaas
```

#### Build & Deploy
```bash 
$ faas-cli up -f pigo-gci.yml
```
or
```bash 
$ faas-cli up -f pigo-nogci.yml
```

### Result
After deploying the OpenFaaS function `pigo-face-detector` will show up in the function list. You just have to hit invoke to run it. At each call, this will return the runtime's id, the function's service time in nanoseconds and garbage collector data.


Sample image used: https://user-images.githubusercontent.com/883386/53553708-ebb88a00-3b46-11e9-9ea8-73c6b7f9dfa1.jpg

## License

Copyright © 2018 Endre Simo

This project is under the MIT License. See the LICENSE file for the full license text.
