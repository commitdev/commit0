---
title: Overview
sidebar_label: Overview
sidebar_position: 1
---


## What is Zero

Zero is an open source tool which makes it quick and easy for startup technical founders and developers to build everything they need to launch and grow high-quality SaaS applications faster and more cost-effectively.

Zero sets up everything you need so you can immediately start building your product.

Zero was created by [Commit](https://commit.dev).
## Why is Zero good for startups

As a technical founder or the first technical hire at a startup, your sole focus is to build the logic for your application and get it into customers’ hands as quickly and reliably as possible. Yet you immediately face multiple hurdles before even writing the first line of code. You’re forced to make many tech trade-offs, leading to decision fatigue. You waste countless hours building boilerplate SaaS features not adding direct value to your customers. You spend precious time picking up unfamiliar tech, make wrong choices that result in costly refactoring or rebuilding in the future, and are unaware of tools and best practices that would speed up your product iteration.

Zero was built by a team of engineers with many years of experience in building and scaling startups. We have faced all the problems you will and want to provide a way for new startups to avoid all those pitfalls. We also want to help you learn about the tech choices we made so your team can become proficient in some of the great tools we have included. The system you get starts small but allows you to scale well into the future when you need to.

Everything built by Zero is yours. After using Zero to generate your infrastructure, backend, and frontend, all the code is checked into your source control repositories and becomes the basis for your new system. We provide constant updates and new modules that you can pull in on an ongoing basis, but you can also feel free to customize as much as you like with no strings attached. If you do happen to make a change to core functionality and feel like contributing it back to the project, we'd love that too!

It's easy to get started, the only thing you'll need is an AWS account. Just enter your AWS CLI tokens or choose your existing profile during the setup process and everything is built for you automatically using infrastructure-as-code so you can see exactly what's happening and easily modify it if necessary.

[Read about the day-to-day experience of using a system set up using Zero][real-world-usage]


## Why is Zero Reliable, Scalable, Performant, and Secure

Reliability: Your infrastructure will be set up in multiple availability zones making it highly available and fault tolerant. All production workloads will run with multiple instances by default, using AWS ELB and Nginx to load balance traffic. All infrastructure is represented with code using [HashiCorp Terraform][terraform] so your environments are reproducible, auditable, and easy to configure.

Scalability: Your services will be running in Kubernetes, with the EKS nodes running in an AWS [Auto Scaling Group][asg]. Both the application workloads and cluster size are ready to scale whenever the need arises. Your frontend assets will be stored in S3 and served from AWS' Cloudfront CDN which operates at global scale.

Security: Properly configured access-control to resources/security groups, using secret storage systems (AWS Secret Manager, Kubernetes secrets), and following best practices provides great security out of the box. Our practices are built on top of multiple security audits and penetration tests. Automatic certificate management using [Let's Encrypt][letsencrypt], database encryption, VPN support, and more means your traffic will always be encrypted. Built-in application features like user authentication help you bullet-proof your application by using existing, tested tools rather than reinventing the wheel when it comes to features like user management and auth.


## What do you get out of the box?
[Read about why we made these technology choices and where they are most applicable.][technology-choices]

[Check out some resources for learning more about these technologies.][learning-resources]

### [Infrastructure][docs-infra]
- Fully configured infrastructure-as-code AWS environment including:
  - VPCs per environment (staging, production) with pre-configured subnets, security groups, etc.
  - EKS Kubernetes cluster per environment, pre-configured with helpful tools like cert-manager, external-dns, nginx-ingress-controller
  - RDS database for your application (Postgres or MySQL)
  - S3 buckets and Cloudfront distributions to serve your assets
- Logging and Metrics collected automatically using either Cloudwatch or Prometheus + Grafana, Elasticsearch + Kibana
- VPN using [Wireguard][wireguard] (Optional)
- User management and Identity / Access Proxy using Ory [Kratos][kratos] and [Oathkeeper][oathkeeper] (Optional)
- Tooling to make it easy to set up secure access for your dev team
- Local/Cloud Hybrid developer environment using Telepresence (Optional)

### [Backend][docs-backend-go] ([Go][docs-backend-go] / [Node.js][docs-backend-nodejs])
- Golang or Node.js example project automatically set up, Dockerized, and deployed to your new Kubernetes cluster
- CI pipeline built with [CircleCI][circleci] or GitHub Actions. Just merge a PR and a deploy will start. Your code will be built and tested, deployed to staging, then prompt you to push to production
- File upload / download support using signed Cloudfront URLs (Optional)
- Email support using [SendGrid][sendgrid] or AWS SES (Optional)
- Notification support for sending and receiving messages in your application (web, mobile, SMS, Email, etc.) (Optional) (In Progress)
- User management integration with Kratos and Oathkeeper - No need to handle login, signup, authentication yourself (Optional)

### [Frontend][docs-frontend-react]
- React example project automatically set up, deployed and served securely to your customers
- CI pipeline built with CircleCI or GitHub Actions. Just merge a PR and a deploy will start. Your code will be built and tested, deployed to staging, then prompt you to push to production
- File upload / download support using signed Cloudfront URLs (Optional)
- User management integration with Kratos - Just style the example login / signup flow to look the way you want (Optional)
- Static site example project using Gatsby to easily make a landing page, also set up with a CI Pipeline using CircleCI (Optional)

<!-- internal links -->
[real-world-usage]: ./real-world-usage
[technology-choices]: ./technology-choices
[learning-resources]: ../reference/learning-resources
[docs-infra]: /docs/modules/aws-eks-stack/
[docs-backend-go]: /docs/modules/backend-go/
[docs-backend-nodejs]: /docs/modules/backend-nodejs/
[docs-frontend-react]: /docs/modules/frontend-react/
<!-- links -->


[git]:      https://git-scm.com
[kubectl]:  https://kubernetes.io/docs/tasks/tools/install-kubectl/
[terraform]:https://www.terraform.io/downloads.html
[jq]:       https://github.com/stedolan/jq
[AWS CLI]:  https://aws.amazon.com/cli/
[acw]:      https://aws.amazon.com/cloudwatch/
[vpc]:      https://aws.amazon.com/vpc/
[iam]:      https://aws.amazon.com/iam/
[asg]:      https://aws.amazon.com/autoscaling/
[zero binary]: https://github.com/commitdev/zero/releases/
[Wget]: https://stackoverflow.com/questions/33886917/how-to-install-wget-in-macos
[and more]: https://github.com/commitdev/zero-aws-eks-stack/blob/master/docs/resources.md
[terraform]: https://terraform.io
[letsencrypt]: https://letsencrypt.org/
[kratos]: https://www.ory.sh/kratos/
[oathkeeper]: https://www.ory.sh/oathkeeper/
[wireguard]: https://wireguard.com/
[circleci]: https://circleci.com/
[sendgrid]: https://sendgrid.com/
[launchdarkly]: https://launchdarkly.com/
