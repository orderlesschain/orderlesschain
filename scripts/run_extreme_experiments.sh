#!/usr/bin/env bash

source "${PROJECT_ABSOLUTE_PATH}"/env
start=$(date +%s)

rm -rf "${PROJECT_ABSOLUTE_PATH}"/configs/certs
mkdir "${PROJECT_ABSOLUTE_PATH}"/configs/certs
if [[ $BUILD_MODE == "local" ]]; then
  cp "${PROJECT_ABSOLUTE_PATH}"/configs/endpoints_local.yml "${PROJECT_ABSOLUTE_PATH}"/configs/endpoints.yml
  cp "${PROJECT_ABSOLUTE_PATH}"/configs/certs_local/* "${PROJECT_ABSOLUTE_PATH}"/configs/certs
else
  cp "${PROJECT_ABSOLUTE_PATH}"/configs/endpoints_remote.yml "${PROJECT_ABSOLUTE_PATH}"/configs/endpoints.yml
  cp "${PROJECT_ABSOLUTE_PATH}"/configs/certs_remote/* "${PROJECT_ABSOLUTE_PATH}"/configs/certs
fi

commitResults(){
    pushd "${PROJECT_ABSOLUTE_PATH}"/orderlesschain-experiments
    git add .
    git commit -m 'auto push'
    git pull origin master
    git push origin master
    popd
}

runExtreme(){
    pushd "${PROJECT_ABSOLUTE_PATH}" || exit


#    for BENCHMARK in sync.scale21 sync.scale22 sync.scale23 sync.scale24 sync.scale25 sync.scale26 sync.scale27 sync.scale28 sync.scale29 sync.scale30 ; do
#            echo "Benchmark $BENCHMARK is executing"
#            go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
#            commitResults
#    done
#
#    for BENCHMARK in sync.scale31 sync.scale32 sync.scale33 sync.scale34 sync.scale35 sync.scale36 sync.scale37 sync.scale38 sync.scale39  sync.scale40 ; do
#                echo "Benchmark $BENCHMARK is executing"
#                go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
#                commitResults
#    done
#
#    for BENCHMARK in sync.scale21 sync.scale22 sync.scale23 sync.scale24 sync.scale25 sync.scale26 sync.scale27 sync.scale28 sync.scale29 sync.scale30 ; do
#            echo "Benchmark $BENCHMARK is executing"
#            go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
#            commitResults
#    done
#
#    for BENCHMARK in sync.scale31 sync.scale32 sync.scale33 sync.scale34 sync.scale35 sync.scale36 sync.scale37 sync.scale38 sync.scale39  sync.scale40 ; do
#                echo "Benchmark $BENCHMARK is executing"
#                go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
#                commitResults
#    done
#
#    for BENCHMARK in sync.scale21 sync.scale22 sync.scale23 sync.scale24 sync.scale25 sync.scale26 sync.scale27 sync.scale28 sync.scale29 sync.scale30 ; do
#            echo "Benchmark $BENCHMARK is executing"
#            go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
#            commitResults
#    done
#
#    for BENCHMARK in sync.scale31 sync.scale32 sync.scale33 sync.scale34 sync.scale35 sync.scale36 sync.scale37 sync.scale38 sync.scale39  sync.scale40 ; do
#                echo "Benchmark $BENCHMARK is executing"
#                go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
#                commitResults
#    done

#     for BENCHMARK in sync.scale5 sync.scale5 sync.scale5  ; do
#                        echo "Benchmark $BENCHMARK is executing"
#                        go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
#                        commitResults
#         done

#     for BENCHMARK in sync.load1 sync.load1 sync.load1  ; do
#                    echo "Benchmark $BENCHMARK is executing"
#                    go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
#                    commitResults
#     done

#    for BENCHMARK in sync.scale1 sync.scale2 sync.scale3 sync.scale4 sync.scale5 sync.scale6 sync.scale7 sync.scale8 sync.scale9 sync.scale10  ; do
#            echo "Benchmark $BENCHMARK is executing"
#            go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
#            commitResults
#    done
#
#    for BENCHMARK in sync.scale11 sync.scale12 sync.scale13 sync.scale14 sync.scale15 sync.scale16 sync.scale17 sync.scale18 sync.scale19 sync.scale20  ; do
#                echo "Benchmark $BENCHMARK is executing"
#                go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
#                commitResults
#    done

#    for BENCHMARK in  sync.byzantine1   sync.byzantine1   sync.byzantine1; do
#        echo "Benchmark $BENCHMARK is executing"
#        go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
#        commitResults
#    done

#    for BENCHMARK in  auction.orderlesschain8 auction.bidl8 vote.orderlesschain8 vote.bidl8; do
#        echo "Benchmark $BENCHMARK is executing"
#        go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
#        commitResults
#    done

#    for BENCHMARK in  auction.orderlesschain1 auction.bidl1 auction.fabric1 auction.fabriccrdt1 auction.synchotstuff1; do
#        echo "Benchmark $BENCHMARK is executing"
#        go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
#        commitResults
#    done
#
#    for BENCHMARK in  vote.orderlesschain1 vote.bidl1 vote.fabric1 vote.fabriccrdt1 vote.synchotstuff1; do
#        echo "Benchmark $BENCHMARK is executing"
#        go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
#        commitResults
#    done

#   for BENCHMARK in auction.orderlesschain8 auction.bidl8 auction.synchotstuff8; do
#        echo "Benchmark $BENCHMARK is executing"
#        go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
#        commitResults
#    done
#
#    for BENCHMARK in  vote.orderlesschain8 vote.bidl8 vote.synchotstuff8; do
#        echo "Benchmark $BENCHMARK is executing"
#        go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
#        commitResults
#    done

#       for BENCHMARK in  auction.bidl7 auction.bidl8 auction.orderlesschain7 auction.orderlesschain8 auction.synchotstuff7 auction.synchotstuff8; do
#            echo "Benchmark $BENCHMARK is executing"
#            go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
#            commitResults
#        done
#
#        for BENCHMARK in  vote.bidl7 vote.bidl8 vote.orderlesschain7 vote.orderlesschain8 vote.synchotstuff7 vote.synchotstuff8; do
#            echo "Benchmark $BENCHMARK is executing"
#            go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
#            commitResults
#        done
#
#     for BENCHMARK in  vote.orderlesschain8  vote.synchotstuff5 vote.synchotstuff6 auction.synchotstuff5 auction.synchotstuff6; do
#              echo "Benchmark $BENCHMARK is executing"
#              go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
#              commitResults
#    done

for BENCHMARK in   vote.orderlesschain8 vote.synchotstuff8 vote.bidl8; do
              echo "Benchmark $BENCHMARK is executing"
              go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
              commitResults
    done

    popd || exit
}

runExtremeTwo(){
    pushd "${PROJECT_ABSOLUTE_PATH}" || exit


  for BENCHMARK in   vote.orderlesschain5 vote.fabric5 vote.fabriccrdt5; do
              echo "Benchmark $BENCHMARK is executing"
              go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
              commitResults
    done

     for BENCHMARK in   auction.orderlesschain5 auction.fabric5 auction.fabriccrdt5; do
                  echo "Benchmark $BENCHMARK is executing"
                  go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
                  commitResults
        done

    popd || exit
}


runByzantine(){
    pushd "${PROJECT_ABSOLUTE_PATH}" || exit


  for BENCHMARK in   sync.byzantine2  sync.byzantine3 sync.byzantine4; do
              echo "Benchmark $BENCHMARK is executing"
              go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
              commitResults
    done

  for BENCHMARK in   sync.byzantine5  sync.byzantine6 sync.byzantine7 sync.byzantine8 sync.byzantine9; do
              echo "Benchmark $BENCHMARK is executing"
              go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
              commitResults
    done

    popd || exit
}


runExtremeThree(){
    pushd "${PROJECT_ABSOLUTE_PATH}" || exit


  for BENCHMARK in   sync.scale30 sync.obj8 sync.endorse8; do
              echo "Benchmark $BENCHMARK is executing"
              go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
              commitResults
    done



    popd || exit
}


##################################################################################

#runExtremeTwo
#
#runExtremeTwo
#
#runExtreme
#
#runExtremeTwo
#
#runExtremeTwo

#
#runExtreme
#
#runExtreme

runByzantine
runByzantine
runByzantine
runByzantine
runByzantine

#runExtremeThree
#runExtremeThree
#runExtremeThree

end=$(date +%s)
echo Experiments executed in $(expr $end - $start) seconds.
