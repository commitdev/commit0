import React from 'react';
import clsx from 'clsx';
import Layout from '@theme/Layout';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import styles from './index.module.scss';
import HomepageFeatures from '../components/HomepageFeatures';
import HomepageTrustedBy from '../components/HomepageTrustedBy';
import HomepageVideo from '../components/HomepageVideo';
import HomepageWhyZero from '../components/HomepageWhyZero';
import HomepageOfferings from '../components/HomepageOfferings';

function HomepageHeader() {
  const {siteConfig} = useDocusaurusContext();

  return (
    <header className={clsx('hero hero--primary', styles.heroBanner)}>
      <p className={styles.hero__subtitle}>{siteConfig.tagline}</p>
    </header>
  );
}

function HomePageCallToAction () {
  return <div className={styles.buttons}>
    <Link
      className="button button-cta button--secondary button--lg"
      to="/docs/zero/getting-started/installation">
      Get Started
    </Link>
  </div>
}

export default function Home() {
  const landingPageOnlyGlobalItemStyle = `
  .navbar  {
    padding: 2.5rem 0 3.5rem;
    box-shadow: none;
    background: linear-gradient(90deg, rgba(15, 16, 17, 0.6) 0%, rgba(1, 2, 66, 0.6) 100%);
  }
  .navbar__inner {
    padding: 0 3rem;
  }
  .navbar__brand img{
    height: 130%;
  }
  .react-toggle{
    display: none;
  }
  `;
  const {siteConfig} = useDocusaurusContext();
  return (
    <Layout
      title={`Build it Fast, Build it Right!`}
      description="Opinionated infrastructure to take you from idea to production on day one!"
      >
      <style>{landingPageOnlyGlobalItemStyle}</style>
      <HomepageHeader />
      <main>
        <HomepageVideo />
        <HomePageCallToAction />
        <HomepageTrustedBy />
        <HomepageFeatures />
        <HomepageWhyZero />
        <HomepageOfferings />
        <HomePageCallToAction />
      </main>
    </Layout>
  );
}
