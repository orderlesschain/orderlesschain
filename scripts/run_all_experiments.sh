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
   # git add .
    #git commit -m 'auto push'
    #git pull origin master
    #git push origin master
    popd
}

############################OrderlessChain Synthetic######################################################

runSyncThroughput(){
    pushd "${PROJECT_ABSOLUTE_PATH}" || exit

    for BENCHMARK in sync.throughput1 sync.throughput2 sync.throughput3 sync.throughput4 sync.throughput5 sync.throughput6 sync.throughput7 sync.throughput8 sync.throughput9 sync.throughput10; do
        for i in {1..1}; do
          echo "Benchmark $BENCHMARK is executing for $i times"
          go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
          commitResults
        done
    done

    popd || exit
}

runSyncCRDTs(){
    pushd "${PROJECT_ABSOLUTE_PATH}" || exit

    for BENCHMARK in sync.crdt1 sync.crdt2 sync.crdt3; do
        for i in {1..1}; do
          echo "Benchmark $BENCHMARK is executing for $i times"
          go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
          commitResults
        done
    done

    popd || exit
}

runSyncEndorse(){
    pushd "${PROJECT_ABSOLUTE_PATH}" || exit

    for BENCHMARK in sync.endorse1 sync.endorse2 sync.endorse3 sync.endorse4 sync.endorse5 sync.endorse6 sync.endorse7 sync.endorse8; do
        for i in {1..1}; do
          echo "Benchmark $BENCHMARK is executing for $i times"
          go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
          commitResults
        done
    done

    popd || exit
}

runSyncObjects(){
    pushd "${PROJECT_ABSOLUTE_PATH}" || exit

    for BENCHMARK in sync.obj1 sync.obj2 sync.obj3 sync.obj4 sync.obj5 sync.obj6 sync.obj7 sync.obj8; do
        for i in {1..1}; do
          echo "Benchmark $BENCHMARK is executing for $i times"
          go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
          commitResults
        done
    done

    popd || exit
}

runSyncOperations(){
    pushd "${PROJECT_ABSOLUTE_PATH}" || exit

    for BENCHMARK in sync.operation1 sync.operation2 sync.operation3 sync.operation4 sync.operation5 sync.operation6 sync.operation7 sync.operation8; do
        for i in {1..1}; do
          echo "Benchmark $BENCHMARK is executing for $i times"
          go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
          commitResults
        done
    done

    popd || exit
}

runSyncOrgs(){
    pushd "${PROJECT_ABSOLUTE_PATH}" || exit

    for BENCHMARK in  sync.orgs1 sync.orgs2 sync.orgs3 sync.orgs4 sync.orgs5 sync.orgs6 sync.orgs7; do
        for i in {1..1}; do
          echo "Benchmark $BENCHMARK is executing for $i times"
          go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
          commitResults
        done
    done

#    for BENCHMARK in  sync.orgs1 sync.orgs3 sync.orgs5 sync.orgs7; do
#            for i in {1..1}; do
#              echo "Benchmark $BENCHMARK is executing for $i times"
#              go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
#              commitResults
#            done
#        done

    popd || exit
}

runSyncWorkload(){
    pushd "${PROJECT_ABSOLUTE_PATH}" || exit

    for BENCHMARK in sync.workload1 sync.workload2 sync.workload3  sync.workload4 sync.workload5 sync.workload6 sync.workload7 sync.workload8 sync.workload9; do
        for i in {1..1}; do
          echo "Benchmark $BENCHMARK is executing for $i times"
          go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
          commitResults
        done
    done

    popd || exit
}

runSyncGossip(){
    pushd "${PROJECT_ABSOLUTE_PATH}" || exit

    for BENCHMARK in sync.gossip1 sync.gossip2 sync.gossip3 sync.gossip4 sync.gossip5 sync.gossip6 sync.gossip7 sync.gossip8; do
        for i in {1..1}; do
          echo "Benchmark $BENCHMARK is executing for $i times"
          go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
          commitResults
        done
    done

    popd || exit
}

runSyncLoadBalancing(){
    pushd "${PROJECT_ABSOLUTE_PATH}" || exit

    for BENCHMARK in sync.load1 sync.load2 sync.load3 sync.load4 sync.load5 sync.load6 sync.load7; do
        for i in {1..1}; do
          echo "Benchmark $BENCHMARK is executing for $i times"
          go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
          commitResults
        done
    done

    popd || exit
}

############################OrderlessChain Auction Vote######################################################

runAuctionOrderlessChain(){
    pushd "${PROJECT_ABSOLUTE_PATH}" || exit

    for BENCHMARK in  auction.orderlesschain1 auction.orderlesschain2 auction.orderlesschain3 auction.orderlesschain4 auction.orderlesschain5 auction.orderlesschain6; do
        for i in {1..1}; do
          echo "Benchmark $BENCHMARK is executing for $i times"
          go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
          commitResults
        done
    done

    popd || exit
}

runAuctionOrderlessChainGaussian(){
    pushd "${PROJECT_ABSOLUTE_PATH}" || exit

    for BENCHMARK in auction.normal.orderlesschain1 auction.normal.orderlesschain2 auction.normal.orderlesschain3 auction.normal.orderlesschain4 auction.normal.orderlesschain5; do
        for i in {1..1}; do
          echo "Benchmark $BENCHMARK is executing for $i times"
          go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
          commitResults
        done
    done

    popd || exit
}

runVoteOrderlessChain(){
    pushd "${PROJECT_ABSOLUTE_PATH}" || exit

    for BENCHMARK in vote.orderlesschain1 vote.orderlesschain2 vote.orderlesschain3 vote.orderlesschain4 vote.orderlesschain5 vote.orderlesschain6; do
        for i in {1..1}; do
          echo "Benchmark $BENCHMARK is executing for $i times"
          go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
          commitResults
        done
    done

    popd || exit
}

runVoteOrderlessChainGaussian(){
    pushd "${PROJECT_ABSOLUTE_PATH}" || exit

    for BENCHMARK in vote.normal.orderlesschain1 vote.normal.orderlesschain2 vote.normal.orderlesschain3 vote.normal.orderlesschain4 vote.normal.orderlesschain5; do
        for i in {1..1}; do
          echo "Benchmark $BENCHMARK is executing for $i times"
          go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
          commitResults
        done
    done

    popd || exit
}

############################Fabric######################################################

runAuctionFabric(){
    pushd "${PROJECT_ABSOLUTE_PATH}" || exit

    for BENCHMARK in auction.fabric1 auction.fabric2 auction.fabric3  auction.fabric4 auction.fabric5 auction.fabric6; do
        for i in {1..1}; do
          echo "Benchmark $BENCHMARK is executing for $i times"
          go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
          commitResults
        done
    done

    popd || exit
}

runAuctionFabricGaussian(){
    pushd "${PROJECT_ABSOLUTE_PATH}" || exit

    for BENCHMARK in auction.normal.fabric1 auction.normal.fabric2 auction.normal.fabric3 auction.normal.fabric4 auction.normal.fabric5; do
        for i in {1..1}; do
          echo "Benchmark $BENCHMARK is executing for $i times"
          go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
          commitResults
        done
    done

    popd || exit
}

runVoteFabric(){
    pushd "${PROJECT_ABSOLUTE_PATH}" || exit

    for BENCHMARK in vote.fabric1 vote.fabric2  vote.fabric3 vote.fabric4 vote.fabric5 vote.fabric6; do
        for i in {1..1}; do
          echo "Benchmark $BENCHMARK is executing for $i times"
          go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
          commitResults
        done
    done

    popd || exit
}

runVoteFabricGaussian(){
    pushd "${PROJECT_ABSOLUTE_PATH}" || exit

    for BENCHMARK in vote.normal.fabric1 vote.normal.fabric2 vote.normal.fabric3 vote.normal.fabric4 vote.normal.fabric5; do
        for i in {1..1}; do
          echo "Benchmark $BENCHMARK is executing for $i times"
          go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
          commitResults
        done
    done

    popd || exit
}

###############################FabricCRDT###################################################

runAuctionFabricCRDT(){
    pushd "${PROJECT_ABSOLUTE_PATH}" || exit

    for BENCHMARK in  auction.fabriccrdt1 auction.fabriccrdt2 auction.fabriccrdt3 auction.fabriccrdt4 auction.fabriccrdt5 auction.fabriccrdt6; do
        for i in {1..1}; do
          echo "Benchmark $BENCHMARK is executing for $i times"
          go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
          commitResults
        done
    done

    popd || exit
}

runAuctionFabricCRDTGaussian(){
    pushd "${PROJECT_ABSOLUTE_PATH}" || exit

    for BENCHMARK in auction.normal.fabriccrdt1 auction.normal.fabriccrdt2 auction.normal.fabriccrdt3 auction.normal.fabriccrdt4 auction.normal.fabriccrdt5; do
        for i in {1..1}; do
          echo "Benchmark $BENCHMARK is executing for $i times"
          go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
          commitResults
        done
    done

    popd || exit
}

runVoteFabricCRDT(){
    pushd "${PROJECT_ABSOLUTE_PATH}" || exit

    for BENCHMARK in vote.fabriccrdt1  vote.fabriccrdt2  vote.fabriccrdt3   vote.fabriccrdt4   vote.fabriccrdt5 vote.fabriccrdt6; do
        for i in {1..1}; do
          echo "Benchmark $BENCHMARK is executing for $i times"
          go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
          commitResults
        done
    done

    popd || exit
}

runVoteFabricCRDTGaussian(){
    pushd "${PROJECT_ABSOLUTE_PATH}" || exit

    for BENCHMARK in vote.normal.fabriccrdt1 vote.normal.fabriccrdt2 vote.normal.fabriccrdt3 vote.normal.fabriccrdt4 vote.normal.fabriccrdt5; do
        for i in {1..1}; do
          echo "Benchmark $BENCHMARK is executing for $i times"
          go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
          commitResults
        done
    done

    popd || exit
}

############################BIDL######################################################

runAuctionBIDL(){
    pushd "${PROJECT_ABSOLUTE_PATH}" || exit

    for BENCHMARK in auction.bidl1 auction.bidl2 auction.bidl3 auction.bidl4 auction.bidl5 auction.bidl6;  do
        for i in {1..1}; do
          echo "Benchmark $BENCHMARK is executing for $i times"
          go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
          commitResults
        done
    done

    popd || exit
}

runVoteBIDL(){
    pushd "${PROJECT_ABSOLUTE_PATH}" || exit

    for BENCHMARK in vote.bidl1 vote.bidl2  vote.bidl3 vote.bidl4 vote.bidl5 vote.bidl6; do
        for i in {1..1}; do
          echo "Benchmark $BENCHMARK is executing for $i times"
          go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
          commitResults
        done
    done

    popd || exit
}

##################################################################################

############################SyncHotStuff######################################################

runAuctionSyncHotStuff(){
    pushd "${PROJECT_ABSOLUTE_PATH}" || exit

    for BENCHMARK in auction.synchotstuff1  auction.synchotstuff2  auction.synchotstuff3  auction.synchotstuff4  auction.synchotstuff5 auction.synchotstuff6 ;  do
        for i in {1..1}; do
          echo "Benchmark $BENCHMARK is executing for $i times"
          go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
          commitResults
        done
    done

    popd || exit
}

runVoteSyncHotStuff(){
    pushd "${PROJECT_ABSOLUTE_PATH}" || exit

    for BENCHMARK in vote.synchotstuff1 vote.synchotstuff2 vote.synchotstuff3 vote.synchotstuff4 vote.synchotstuff5 vote.synchotstuff6 ; do
        for i in {1..1}; do
          echo "Benchmark $BENCHMARK is executing for $i times"
          go run ./cmd/client -coordinator=true -benchmark=${BENCHMARK}
          commitResults
        done
    done

    popd || exit
}


##################################################################################

runSyncOrderlessChainExperiments(){
    runSyncThroughput
    runSyncCRDTs
    runSyncEndorse
    runSyncObjects
    runSyncOperations
    runSyncOrgs
    runSyncWorkload
    runSyncGossip
    runSyncLoadBalancing
}

runAuction(){
   runAuctionOrderlessChain
   runAuctionFabric
   runAuctionFabricCRDT
   runAuctionBIDL
   runAuctionSyncHotStuff
}

runVote(){
  runVoteOrderlessChain
  runVoteFabric
  runVoteFabricCRDT
  runVoteBIDL
  runVoteSyncHotStuff
}

runAuctionNormal(){
   runAuctionOrderlessChainGaussian
   runAuctionFabricGaussian
   runAuctionFabricCRDTGaussian
}

runVoteNormal(){
   runVoteOrderlessChainGaussian
   runVoteFabricGaussian
   runVoteFabricCRDTGaussian
}

runSyncOrderlessChainExperiments

runVote

runAuction


runSyncOrderlessChainExperiments

runAuction

runVote

runSyncOrderlessChainExperiments

runAuction

runVote


runSyncOrderlessChainExperiments

runAuction

runVote

runSyncOrderlessChainExperiments

runAuction

runVote



end=$(date +%s)
echo Experiments executed in $(expr $end - $start) seconds.
