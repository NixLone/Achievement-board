import type { ReactNode } from "react";

export function ListItem({
  icon,
  title,
  subtitle,
  trailing,
  onClick
}: {
  icon: string;
  title: string;
  subtitle?: string;
  trailing?: ReactNode;
  onClick?: () => void;
}) {
  return (
    <div className="list-item" onClick={onClick} role={onClick ? "button" : undefined}>
      <div className="list-item__icon">{icon}</div>
      <div className="list-item__body">
        <p>{title}</p>
        {subtitle && <span className="muted">{subtitle}</span>}
      </div>
      {trailing && <div className="list-item__trailing">{trailing}</div>}
    </div>
  );
}
