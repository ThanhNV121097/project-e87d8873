import { T } from "../editable";

/**
 * A HEADING THAT CHANGES WEIGHT INSIDE THE LINE: a light run, a bold run, a
 * light run. It is this look's one typographic device — there is no second
 * face and no colour in the type — and the bold run is where the eye lands.
 *
 * Three keys under `k` (light, bold, tail), each a run the owner edits in the
 * Studio. A run left empty draws nothing.
 */
export default function Weights({ k, className = "" }: { k: string; className?: string }) {
  return (
    <span className={className}>
      <T k={`${k}.light`} className="font-light" />{" "}
      <T k={`${k}.bold`} className="font-bold" />{" "}
      <T k={`${k}.tail`} className="font-light" />
    </span>
  );
}
