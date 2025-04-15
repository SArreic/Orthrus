package request

import (
    "crypto/sha256"
    pb "github.com/Hanzheng2021/orthrus/protobufs"
)

func computeDigest(req *pb.ClientRequest) []byte {
    h := sha256.New()
    h.Write([]byte{
        byte(req.RequestId.ClientId >> 8),
        byte(req.RequestId.ClientId),
        byte(req.RequestId.ClientSn >> 8),
        byte(req.RequestId.ClientSn),
    })
    h.Write(req.Payload)
    return h.Sum(nil)
}