import { Link } from "react-router-dom";
import { T, useContent } from "../editable";
import { Picture } from "../picture";
import { SignInForm } from "../components/admin/SignInPage";
import Icon from "../components/Icon";
import Logo from "../components/Logo";
import Weights from "../components/Weights";

/**
 * The way into the shop's own side: a riding photograph down the left half,
 * running to the screen's edge, and at the right the mark, a heading that
 * changes weight like every head on the site, and the form.
 */
export default function AdminSignIn() {
  const photo = useContent("pages.admin.signin.image");
  return (
    <div className="flex min-h-screen flex-col bg-ground font-body text-ink lg:flex-row">
      {photo && (
        <div className="relative h-56 overflow-hidden sm:h-80 lg:h-auto lg:w-1/2">
          <Picture src={photo} alt="" width={850} height={1280} loading="eager" className="absolute inset-0 h-full w-full object-cover" />
        </div>
      )}
      <div className="flex flex-1 items-center px-[var(--gutter)] py-12 lg:py-16">
        <div className="w-full max-w-md">
          <Link to="/" className="inline-flex"><Logo /></Link>
          <h1 className="mt-12 text-[length:var(--size-h1)] leading-[1.1] text-ink"><Weights k="pages.admin.signin.heading" /></h1>
          <T k="pages.admin.signin.description" as="p" className="mt-4 text-sm text-ink-soft" />
          <div className="mt-10"><SignInForm to="/admin" note={useContent("pages.admin.signin.note")} /></div>
          <Link to="/" className="mt-10 inline-flex items-center gap-2 text-sm text-ink-soft hover:text-ink">
            <Icon name="back" size={16} /><T k="pages.admin.signin.back" />
          </Link>
        </div>
      </div>
    </div>
  );
}
