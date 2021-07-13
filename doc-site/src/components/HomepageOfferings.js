import React, { useState } from 'react';
import styles from './HomepageOfferings.module.scss';

const offerings = [
  {
    logo: require('../../static/img/icons/offering-infra.svg').default,
    label: 'Infrastructure',
    tools: [
      {
        name: 'Terraform',
        logo: 'img/tools/terraform.png',
        url: 'https://terraform.io',
      },
      {
        name: 'Kubernetes',
        logo: 'img/tools/kubernetes.png',
        url: 'https://kubernetes.io/',
        noCrop: true,
      },
      {
        name: 'Amazon Web Services',
        logo: 'img/tools/aws.png',
        url: 'https://aws.amazon.com/',
      },
      {
        name: 'Cert Manager',
        logo: 'img/tools/cert-manager.png',
        url: 'https://cert-manager.io/docs/',
      },
      {
        name: 'External DNS',
        logo: 'img/tools/external-dns.png',
        url: 'https://github.com/kubernetes-sigs/external-dns',
      },
      {
        name: 'Wireguard',
        logo: 'img/tools/wireguard.png',
        url: 'https://www.wireguard.com/',
      },
      {
        name: 'Prometheus',
        logo: 'img/tools/prometheus.png',
        url: 'https://prometheus.io/',
      },
      {
        name: 'Grafana',
        logo: 'img/tools/grafana.png',
        url: 'https://grafana.com/',
      },
    ]
  },
  {
    logo: require('../../static/img/icons/offering-cicd.svg').default,
    label: 'CI/CD',
    tools: [
      {
        name: 'Github Actions',
        logo: 'img/tools/github-actions.svg',
        noCrop: true,
        url: 'https://google.ca',
      },
      {
        name: 'CircleCI',
        logo: 'img/tools/circleci.png',
        url: 'https://circleci.com',
      },
      {
        name: 'Docker',
        logo: 'img/tools/docker.png',
        url: 'https://docker.com/',
      },
      {
        name: 'AWS ECR',
        logo: 'img/tools/ecr.png',
        url: 'https://aws.amazon.com/ecr/',
      },
    ]
  },
  {
    logo: require('../../static/img/icons/offering-frontend.svg').default,
    label: 'FRONTEND',
    tools: [
      {
        name: 'React js',
        logo: 'img/tools/react.png',
        url: 'https://reactjs.org',
      },
      {
        name: 'AWS S3',
        logo: 'img/tools/s3.png',
        url: 'https://aws.amazon.com/s3/',
      },
      {
        name: 'AWS Cloudfront',
        logo: 'img/tools/cloudfront.png',
        url: 'https://aws.amazon.com/cloudfront/',
      },
      {
        name: 'ECMAScript 2018',
        logo: 'img/tools/js.png',
        url: 'https://www.w3schools.com/js/js_2018.asp',
      },
    ]
  },
  {
    logo: require('../../static/img/icons/offering-backend.svg').default,
    label: 'BACKEND',
    tools: [
      {
        name: 'Golang',
        logo: 'img/tools/golang.png',
        url: 'https://golang.org',
      },
      {
        name: 'Node.js',
        logo: 'img/tools/nodejs.png',
        url: 'https://nodejs.org',
        noCrop: true,
      },
      {
        name: 'Open ID Connect',
        logo: 'img/tools/openid.png',
        url: 'https://openid.net/connect/',
      },
      {
        name: 'Ory Kratos & Oathkeeper',
        logo: 'img/tools/ory.png',
        url: 'https://www.ory.sh/kratos/docs/',
      },
      {
        name: 'Telepresence',
        logo: 'img/tools/telepresence.png',
        url: 'https://www.telepresence.io/',
        noCrop: true,
      },
      {
        name: 'Stripe',
        logo: 'img/tools/stripe.png',
        url: 'https://stripe.com',
        noCrop: true,
      },
    ]
  },
]

const Offerings = ({data, active, clickHandler}) => (
  <div className={styles.offering_box}>
    <div className={styles.left_box}>
      {
        data.map((i, idx) =>
          <Discipline
            key={idx}
            logo={i.logo}
            label={i.label}
            clickHandler={clickHandler}
            active={i.label == active}
          />
        )
      }
    </div>

    <div className={styles.right_box}>
      <ToolBox
        data={ data.find((i) => i.label == active).tools }
      />
    </div>
  </div>
)

const Discipline = ({logo: LogoSvg, label, clickHandler, active}) => (
  <div
    className={`${styles.discipline} ${active && styles.discipline_active}`}
    onClick={() => clickHandler({ active: label})}
  >
    <LogoSvg className={styles.logo} alt="logo" />
    <h3 className={styles.discipline_name}>{label}</h3>
  </div>
)

const ToolBox = ({data}) => <ul>
  {
    data.map((tool, idx) =>
      <Tool key={idx} tool={tool} idx={idx} />
    )
  }
</ul>

const Tool = ({tool, idx}) => (<li key={`tool-${idx}`}>
  <a href={tool.url} target="_blank">
    <img src={tool.logo} className={tool.noCrop && styles["no-crop"]} />
    <p>{tool.name}</p>
  </a>
</li>)

export default function FeaturedOfferings() {
  const title = "What do you get out of the box ?"
  const [state, setState] = useState({
    active: "Infrastructure",
  })

  return <div className={`${styles.offerings_container} featured-sections`}>
    <h2>{title}</h2>
    <Offerings data={offerings} active={state.active} clickHandler={setState} />
  </div>
}
