# Go: Virtual Machine Detection
A little Go script for detecting if the code is run in a Virtual Machine. Works on some common identifiers.

## How does it work?
The Script tests a few low hanging fruits. Usually these are settings activated to make live easier on a Virtual Machine, but that also can be tested with the ``user`` access rights.

### Multiple Screens
Usually a good indication for a non VM are multiple screens as an output source. Most VM Environments can have multiple screens but in practice it's rarely used. Not to mention that for using Multiple screens on VirtualBox, the Guest Additions must be installed (and we can check that ;))

### Wired Aspect Ratio
A non standard Aspect Ratio is not a 100% a VM but a relatively good indication for a VM. The Script Tests for the most common ones (16:10, 16:9, 4:3, 3:2 and 32:9). Also to resize the window in VirtualBox the Guest Additions need to be installed.

### Long or Short Uptime
As a result of VM usage, the uptime is either very short or very long. The Script tests for <10 minutes and >2 days.

### Size of Main Drive
If the Main Drive is <200 GB. As there is no commercial incentive to Build Computers that have less than that as Main Disk Space. It is a good indication for a VM, because usually the RAM is increased before the Disk Space is increased in normal VM operation.

### Suspect RAM / Main Drive Ratio
Based on the Test above there is also a check in the Script to test if the RAM/Disk Ratio is >7.5%. This is almost a dead giveaway of a VM.

### RAM Size an Even Number
The Script tests if the RAM size is an even number. On most systems the RAM Size is usually even. This is due to most commercial RAM modules being a Power of 2. And a combination of such modules is always an even number.

### RAM Size in GB
Check if the size of the RAM is in Gigabytes. If this is not the case then it's a System with a 512 MB RAM module but that is also pretty sus. Also If the Virtual Environments RAM is wrongly configured, then this test will flag.

### Checking of Kernel Modules
Usually to make it easier to use Virtual Machines some Kernel Modules are installed. If we are running on Linux we can even check them with only ``user`` rights. For VirtualBox we test the ``vboxguest`` module. On VMWare we can test for the ``vme_fake`` module (I'm not 100% sure on this one).

### CD-ROM Drive
Usually most VM's have a CD-ROM Drive automatically installed. This isn't a bulletproof way of finding VMs as some desktop and laptop computers still use drives, but in combination with other factors is really helpful.

## Some Improvements
* Sleeper thread to check if uptime and process runtime deltas differ. To check for process debugging.
* Sleeper thread to check every few minutes if variables change.
* Check if the mouse teleports
* Put everything into an easy to use GO Module