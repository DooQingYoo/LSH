# 添加了pbc(https://crypto.stanford.edu/pbc/）库

FROM hyperledger/fabric-ccenv:1.4.1
COPY ./pbc /opt/gopath/src/pbc
RUN cd /opt/gopath/src/pbc \
 && apt-get update \
 && apt-get install libgmp-dev -y \
 && apt-get install build-essential flex bison -y \
 && ./configure \
 && make \
 && sudo make install \
 && ldconfig \
 && rm -r /opt/gopath/src/pbc