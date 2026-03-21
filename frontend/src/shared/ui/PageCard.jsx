import { ArrowRight } from "lucide-react";
import { Link } from "react-router-dom";

export default function PageCard({
  title,
  description,
  links = [],
  embedded = false,
  wide = false,
  children,
}) {
  const cardClassName = [
    "placeholder-card",
    embedded ? "placeholder-card--embedded" : "",
    wide ? "placeholder-card--wide" : "",
  ]
    .filter(Boolean)
    .join(" ");

  const content = (
    <section className={cardClassName}>
      <header className="placeholder-card__header">
        <div className="placeholder-card__copy">
          <h1 className="placeholder-card__title">{title}</h1>
          <p className="placeholder-card__description">{description}</p>
        </div>

        {links.length ? (
          <nav className="quick-links" aria-label="Quick links">
            {links.map((item) => (
              <Link key={item.to} to={item.to} className="quick-links__item">
                <span>{item.label}</span>
                <ArrowRight size={16} strokeWidth={2.1} />
              </Link>
            ))}
          </nav>
        ) : null}
      </header>

      {children}
    </section>
  );

  if (embedded) {
    return content;
  }

  return <main className="screen">{content}</main>;
}
