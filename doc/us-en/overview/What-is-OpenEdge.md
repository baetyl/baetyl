# What is Baetyl

**[Baetyl](https://baetyl.io) is an open edge computing framework that extends cloud computing, data and service seamlessly to edge devices.** It can provide temporary offline, low-latency computing services, and include device connect, message routing, remote synchronization, function computing, video access pre-processing, AI inference, device resources report etc. The combination of Baetyl and the **Cloud Management Suite** of [BIE](https://cloud.baidu.com/product/bie.html)(Baidu IntelliEdge) will achieve cloud management and application distribution, enable applications running on edge devices and meet all kinds of edge computing scenario.

About architecture design, Baetyl takes **modularization** and **containerization** design mode. Based on the modular design pattern, Baetyl splits the product to multiple modules, and make sure each one of them is a separate, independent module. In general, Baetyl can fully meet the conscientious needs of users to deploy on demand. Besides, Baetyl also takes containerization design mode to build images. Due to the cross-platform characteristics of docker to ensure the running environment of each operating system is consistent. In addition, **Baetyl also isolates and limits the resources of containers**, and allocates the CPU, memory and other resources of each running instance accurately to improve the efficiency of resource utilization.

More about Baetyl, please visit:

- [Baetyl advantages](./Baetyl-advantages.md)
- [Baetyl design](./Baetyl-design.md)
- [Baetyl open framework](./Baetyl-open-framework.md)
- [Baetyl customize extension](./Baetyl-extension.md)
- [Baetyl security](./Baetyl-security.md)
- [Baetyl control](./Baetyl-control.md)