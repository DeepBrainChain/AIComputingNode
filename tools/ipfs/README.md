# IPFS

Upload files to or download files from IPFS nodes.

```shell
$ ./ipfs -node /ip4/192.168.1.159/tcp/4002 -upload ./test.png
2024/04/09 17:01:35 Upload file /ipfs/QmStNRFDoBzuEn4g7wWXV7UEFXtXGBf4SgJjzmjjDvD1Hb
2024/04/09 17:01:35 File uploaded successfully.
$ ./ipfs -node /ip4/192.168.1.159/tcp/4002 -download QmStNRFDoBzuEn4g7wWXV7UEFXtXGBf4SgJjzmjjDvD1Hb -save ./test.png
2024/04/09 17:03:22 File downloaded successfully.
```

When we upload a file to an IPFS node (such as `/ip4/192.168.1.159/tcp/4002` in the above example), a CID identification (such as `QmStNRFDoBzuEn4g7wWXV7UEFXtXGBf4SgJjzmjjDvD1Hb` in the above example) will be returned, and then we can view the file in the browser through `http://192.168.1.159:4040/ipfs/cid`, or even view it on the public Internet through `https://ipfs.io/ipfs/cid/`.
