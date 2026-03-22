import { ArrowRight } from "lucide-react";
import { motion, useReducedMotion } from "motion/react";
import { Link } from "react-router-dom";
import {
  createFadeMove,
  createRevealContainer,
} from "../lib/motion";

export default function PageCard({
  title,
  description,
  links = [],
  embedded = false,
  wide = false,
  children,
}) {
  const reducedMotion = useReducedMotion();
  const sectionVariants = createRevealContainer(reducedMotion, {
    staggerChildren: 0.08,
    delayChildren: 0.04,
  });
  const headerVariants = createFadeMove(reducedMotion, {
    axis: "y",
    distance: 18,
    scale: 0.992,
  });
  const itemVariants = createFadeMove(reducedMotion, {
    axis: "y",
    distance: 12,
    scale: 0.996,
  });
  const cardClassName = [
    "placeholder-card",
    embedded ? "placeholder-card--embedded" : "",
    wide ? "placeholder-card--wide" : "",
  ]
    .filter(Boolean)
    .join(" ");

  const content = (
    <motion.section
      animate="visible"
      className={cardClassName}
      initial="hidden"
      variants={sectionVariants}
    >
      <motion.header
        className="placeholder-card__header"
        variants={headerVariants}
      >
        <motion.div className="placeholder-card__copy" variants={itemVariants}>
          <h1 className="placeholder-card__title">{title}</h1>
          <p className="placeholder-card__description">{description}</p>
        </motion.div>

        {links.length ? (
          <motion.nav
            className="quick-links"
            aria-label="Quick links"
            variants={sectionVariants}
          >
            {links.map((item) => (
              <motion.div key={item.to} variants={itemVariants}>
                <Link to={item.to} className="quick-links__item">
                  <span>{item.label}</span>
                  <ArrowRight size={16} strokeWidth={2.1} />
                </Link>
              </motion.div>
            ))}
          </motion.nav>
        ) : null}
      </motion.header>

      <motion.div variants={headerVariants}>{children}</motion.div>
    </motion.section>
  );

  if (embedded) {
    return content;
  }

  return <main className="screen">{content}</main>;
}
