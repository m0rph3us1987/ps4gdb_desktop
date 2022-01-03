# PS4GDB

PS4GDB consists of two components. The first component is the gdbstub running on your ps4 and the second one is ps4gdb_desktop.

## gdbstub

The gdbstub is integrated into Mira as a plugin. PS4GDB has implemented all features except x86 debug registers. I never needed them, but since it's
part of Mira and Mira is open source, feel free to add features you miss. When the stub is loaded it creates a new kernel process, this process will then
listen for incoming connections on port **8146**.

### What is port 8146 used for?
We somehow need to tell the ps4 what process we want to debug, and since hardcoding pids into PS4GDB is not an option, we need a way to communicate and tell
PS4GDB what it is supposed to do for us. This communication is done over port **8146**. PS4GDB spawns a little RPC server, which accepts a couple of commands.
Here is where the second component ps4gdb_desktop comes into play.

## ps4gdb_desktop
ps4gdb_desktop is the component running on your PC able to communicate with PS4GDB. It's written in Go and the protocol behind it is trivial.
It gives you the ability to read a list of processes running on your ps4, attach to a certain pid or kill the PS4GDB kernel process.


To get a list of processes running on your ps4, you just need to run the following command (change the ip to your ps4 ip)
```
ps4gdb_desktop 192.168.0.102:8146 get-pids
```
This should give you a result similiar to this:
![ps4gdb-desktop-getpids](https://i.postimg.cc/SRMxB9nf/ps4gdb-desktop-getpids.png)

This pictures shows the processes running on the ps4 and the corresponding pids.
If now for example we want to debug SceRemotePlay we would execute following command:
```
ps4gdb_desktop 192.168.0.102:8146 attach 95
```

This is basically all you need to know about ps4gdb_desktop. In the next section I will show how to connect to PS4GDB with gdb and debug
an application.

## Debug Playroom
In this section i will show you how to debug playroom. What you see here applies to any other **userland** application. I assume Mira is already
loaded at this point. If you have the possibility to disable userland ASLR do it, because it will make your life much easier.

There are a couple of commands we need to pass to gdb in every debug session, to avoid this i created a file in my home directory thats called
ps4.source, the content looks like this:

```
set architecture i386:x86-64
target remote 192.168.0.2:8846
```

**You must replace my ps4 ip with yours in the second line.**


- Start Playroom

- When playroom is running we need to find out it's pid, so we ask ps4gdb_desktop to give us the pid list:
```
ps4gdb_desktop 192.168.0.102:8146 get-pids
```
This is what the result looks like:

![ps4gdb-desktop-playroom-pid](https://i.postimg.cc/kgBJH445/ps4gdb-desktop-playroom-pid.png)

eboot.bin pid 112 is our candidate. In your case the pid might be different.

- Next we attach to pid 112 with the following command:

```
ps4gdb_desktop 192.168.0.102:8146 attach 112
```

As soon as you issue the command, Playroom will freeze, this is normal. In kernel log you should see that PS4GDB has now taken control.

```
[handle_exception] gdb_stub: handle exception start...
[handle_exception] remcomOutBuffer allocated at 0xffff9fbf37108000
[print_register_info] received interrupt 01 - errorCode: 0x0
[print_register_info] RAX: 0x0000000000000004		RBX:  0x000000088005aa80
[print_register_info] RCX: 0x000000088005aa80		RDX:  0x0000000000000006
[print_register_info] RSI: 0x0000000000000008		RDI:  0x000000088005ab24
[print_register_info] RBP: 0x00000007ed761850		RSP:  0x00000007ed7617a8
[print_register_info] R8:  0x0000000000000000		R9:   0x000000000102023d
[print_register_info] R10: 0x000000000515ca11		R11:  0x00000000000002d0
[print_register_info] R12: 0x0000000000000000		R13:  0x000000088005ab20
[print_register_info] R14: 0x0000000000024c5a		R15:  0x0000000000000000
[print_register_info] RIP: 0x0000000800002c4c		FLAGS:0x0000000000000247
[print_register_info] CS:  0x0000000000000043		SS:   0x000000000000003b
[print_register_info] DS:  0x000000000000003b		ES:   0x000000000000003b
[print_register_info] FS:  0x0000000000000013		GS:   0x000000000000001b
[handle_exception] gdb_stub: Entering main loop...
[getpacket] remcomInBuffer allocated at 0xffff9fbf07ee0000
```

- Now we start gdb

```
gdb
```

- Next we load our ps4.source file created at the beginning using the source command

```
source ps4.source
```

After issuing the command, gdb connects to PS4GDB and you can start debugging like you would debug any
other PC application. This is what it looks like for me:

![ps4gdb-desktop-gdb](https://i.postimg.cc/QdD0MB26/ps4gdb-desktop-gdb.png)

gdb looks like this because i use [Andrea Cardacis gdb-dashboard](https://github.com/cyrus-and/gdb-dashboard)

Have fun and happy debugging.

## Tips

- When you finish debugging, always detach from the process. You can detach by issuing command **q** in gdb.
- Beforing resuming execution, make sure you have first set some breakpoints. If after resuming execution
  your breakpoints don't trigger, you wont be able to detach from the process anymore.
