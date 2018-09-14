Controller Development Guidelines
=================================

Below you will find some guidelines to keep in mind when developing a controller.

## Idempotence

__Idempotence__: _The property of certain operations in mathematics and computer science that they can be applied multiple times without changing the result beyond the initial application<sup>[1]</sup>_

### In a Controller
Considering the idempotence of the actions we take against any shared resource allows for the running of multiple instances of a controller. This has strong positive effects on uptime. Ostensibly, the reason to have more than one instance of a controller running is for the sake of resilience, but it also simplifies rolling upgrades: An instance with a newer version could be started while the older instance is torn down without any downtime.

#### Example: The API Server is Trustworthy

![a tale of n controllers](https://user-images.githubusercontent.com/1194436/31556517-3d923c42-affa-11e7-8fa7-e8ad623570d6.png)

One of the reasons for writing idempotent controllers comes from the manner in which events are delivered.  The API Server uses SSE as an event delivery mechanism, which can only work by delivering events to all subscribers of some source.  In order to mitigate this in an idempotent manner, we can do two things, each of which hinges on a guarantee from the API Server that it is sequentially consistent<sup>[2]</sup>.

1. https://en.wikipedia.org/wiki/Idempotence
2. https://en.wikipedia.org/wiki/Sequential_consistency
