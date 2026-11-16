import { Link } from "react-router-dom";
import { T, useContent } from "../editable";
import { Picture } from "../picture";
import { SignInForm } from "../components/admin/SignInPage";
import Icon from "../components/Icon";
import Logo from "../components/Logo";
import Weights from "../components/Weights";

/**
 * The way into the shop's own side. Down the left half, running to the
 * screen's edges: a riding photograph, and under it the shop's navy — the
 * colour of every button on the site, here as a panel that carries the
 * welcome. At the right, the mark and the form. On a phone the same order,
 * top to bottom: photograph, navy welcome, form.
 */
export default function AdminSignIn() {
  const photo = useContent("pages.admin.signin.image");
  return (
    <div className="flex min-h-screen flex-col bg-ground font-body text-ink lg:flex-row">
      <div className="flex flex-col lg:w-1/2">
        {photo && (
          <div className="relative h-48 overflow-hidden sm:h-72 lg:h-auto lg:flex-1">
            <Picture src={photo} alt="" width={850} height={1280} loading="eager" className="absolute inset-0 h-full w-full object-cover" />
          </div>
        )}
        <div className="bg-accent px-[var(--gutter)] py-8 text-accent-ink lg:px-12 lg:py-12">
          <h1 className="text-[length:var(--size-h1)] leading-[1.1] text-accent-ink"><Weights k="pages.admin.signin.heading" /></h1>
          <T k="pages.admin.signin.description" as="p" className="mt-3 max-w-[40ch] text-sm opacity-80" />
        </div>
      </div>
      <div className="flex flex-1 items-center px-[var(--gutter)] py-12 lg:py-16">
        <div className="w-full max-w-md">
          <Link to="/" className="inline-flex"><Logo /></Link>
          <div className="mt-10"><SignInForm to="/admin" note={useContent("pages.admin.signin.note")} /></div>
          <Link to="/" className="mt-10 inline-flex items-center gap-2 text-sm text-ink-soft hover:text-ink">
            <Icon name="back" size={16} /><T k="pages.admin.signin.back" />
          </Link>
        </div>
      </div>
    </div>
  );
}
