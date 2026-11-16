import { type ReactNode } from "react";
import { Navigate, Outlet, Route, Routes } from "react-router-dom";
import ContactDock from "./components/ContactDock";
import { Motion } from "./motion";
import SiteFooter from "./components/SiteFooter";
import SiteHeader from "./components/SiteHeader";
import { useSession } from "./lib/session";
import AdminKinds from "./pages/AdminKinds";
import AdminOverview from "./pages/AdminOverview";
import AdminProduct from "./pages/AdminProduct";
import AdminProducts from "./pages/AdminProducts";
import AdminSignIn from "./pages/AdminSignIn";
import Bag from "./pages/Bag";
import Checkout from "./pages/Checkout";
import Contact from "./pages/Contact";
import Home from "./pages/Home";
import Product from "./pages/Product";
import Shop from "./pages/Shop";

/**
 * Two sites in one address: the shop a customer visits, and the side the
 * shop runs itself from.
 *
 * They do not share a shell. A customer's page sits between the search
 * header and the footer, inside the page's measure; a management page sits in
 * <DashboardLayout>, which each of them brings with it, and never under the
 * shop's header. So the shop's routes are children of one layout route and
 * the management routes are not.
 */
function Shell() {
  return (
    <div className="min-h-screen bg-ground font-body text-ink">
      {/* The motion the owner chose, on every page (src/motion). */}
      <Motion />
      <SiteHeader />
      <main className="mx-auto max-w-page px-[var(--gutter)]"><Outlet /></main>
      <SiteFooter />
      <ContactDock />
    </div>
  );
}

/**
 * The management area's door. Through the scaffold's own session
 * (src/lib/session.ts) and nothing else: the Studio and the walk sign a
 * visitor in by writing THAT session, and a page gated on any other is drawn
 * and walked signed out.
 */
function SignedIn({ children }: { children: ReactNode }) {
  return useSession() ? <>{children}</> : <Navigate to="/admin/sign-in" replace />;
}

export default function App() {
  return (
    <Routes>
      <Route element={<Shell />}>
        <Route path="/" element={<Home />} />
        <Route path="/shop" element={<Shop />} />
        <Route path="/shop/:id" element={<Product />} />
        <Route path="/bag" element={<Bag />} />
        <Route path="/checkout" element={<Checkout />} />
        <Route path="/contact" element={<Contact />} />
      </Route>
      <Route path="/admin/sign-in" element={<AdminSignIn />} />
      <Route path="/admin" element={<SignedIn><AdminOverview /></SignedIn>} />
      <Route path="/admin/products" element={<SignedIn><AdminProducts /></SignedIn>} />
      <Route path="/admin/products/:id" element={<SignedIn><AdminProduct /></SignedIn>} />
      <Route path="/admin/kinds" element={<SignedIn><AdminKinds /></SignedIn>} />
    </Routes>
  );
}
