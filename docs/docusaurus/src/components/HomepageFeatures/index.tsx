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
    title: 'Language Agnostic',
    description: (
      <>
        Supports any language with a Tree-sitter grammar. Parse and generate code for Python, TypeScript, Go, and more using a unified interface.
      </>
    ),
  },
  {
    title: 'Lua Scripting',
    description: (
      <>
        Use the full power of Lua to define your code generation logic. No complex DSLs to learn—just standard Lua with powerful bindings.
      </>
    ),
  },
  {
    title: 'AST-Aware',
    description: (
      <>
        Access detailed source code structure, not just regex-based matching. Query and manipulate the AST to perform precise code transformations.
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
