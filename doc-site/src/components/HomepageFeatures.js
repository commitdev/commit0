import React from 'react';
import clsx from 'clsx';
import styles from './HomepageFeatures.module.scss';

const FeatureList = [
  {
    title: 'Reliable',
    Svg: require('../../static/img/icons/attr-reliable.svg').default,
    description: (
      <>
        Your infrastructure will be highly available and fault tolerant. 
        Production workloads will be self-healing, with all traffic load 
        balanced to multiple instances of your application. All the 
        infrastructure is represented with code to be reproducible, 
        auditable, and easy to configure.
      </>
    ),
  },
  {
    title: 'Scalable',
    Svg: require('../../static/img/icons/attr-scalable.svg').default,
    description: (
      <>
        Everything in your system will scale automatically based on the needs 
        of your application. For frontend assets, using a CDN will ensure global 
        scale, so you don’t need to worry about it.
      </>
    ),
  },
  {
    title: 'Secure',
    Svg: require('../../static/img/icons/attr-secure.svg').default,
    description: (
      <>
        All your systems will follow security best practices backed up
        by multiple security audits and penetration tests, and will be 
        properly configured to make sure access is controlled to 
        private networks, secrets, and data. Built-in application 
        features help you bullet-proof your application by using 
        existing, tested tools rather than reinventing the wheel.
      </>
    ),
  },
  {
    title: 'Modular',
    Svg: require('../../static/img/icons/attr-modular.svg').default,
    description: (
      <>
        Everything built by Zero is yours. After Zero generates your 
        infrastructure, backend, and frontend, all the code is checked 
        into your source control repositories and becomes the basis for 
        your new system. You can customize as much as you like. We 
        provide constant updates and new modules.
      </>
    ),
  },
];

const Feature = ({Svg, title, description}) => (
  <div className={clsx('col col--5') + " feature"}>
    <div className="text--center">
      <Svg className={styles.featureSvg} alt={title} />
    </div>
    <div className="text--center padding-horiz--md">
      <h3>{title}</h3>
      <p className="description">{description}</p>
    </div>
  </div>
)

export default function HomepageFeatures() {
  return (<>
    <section className={`${styles.features} featured-sections`}>

      <h2 className={styles.title}>Building something <strong>fast</strong> doesn't mean you can't also build it <strong>right</strong></h2>
      <div className={`${styles.row} row`}>
        {FeatureList.map((props, idx) => (
          <Feature key={idx} {...props} />
        ))}
      </div>
    </section>
  </>)
}
