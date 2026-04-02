import type {ReactNode} from 'react';
import clsx from 'clsx';
import Heading from '@theme/Heading';
import styles from './styles.module.css';

type FeatureItem = {
  title: string;
  // Svg: React.ComponentType<React.ComponentProps<'svg'>>; // Removed SVG type
  description: ReactNode;
};

const FeatureList: FeatureItem[] = [
  {
    title: 'Code And Spec Inputs',
    description: (
      <>
        Parse source code with Tree-sitter, work with JSON, YAML, TOML, and XML, and build generation workflows from the files your project already treats as source of truth.
      </>
    ),
  },
  {
    title: 'Programmable Rules',
    description: (
      <>
        Use Lua and JQ to define reusable generation rules. Agents can help maintain the rules, while cogeni executes them deterministically.
      </>
    ),
  },
  {
    title: 'Derived Artifacts',
    description: (
      <>
        Generate docs, SDKs, CLIs, schemas, config, and synchronized file sections from a shared runtime with dependency-aware execution.
      </>
    ),
  },
];

function Feature({title, description}: FeatureItem) {
  return (
    <div className={clsx('col col--4')}>
      <div className="text--center padding-horiz--md">
        <Heading as="h3">{title}</Heading>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures(): ReactNode {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
