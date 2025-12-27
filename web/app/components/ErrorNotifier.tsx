import { useState } from "react";
import { useErrors } from "~/context/error";

function ErrorNotifier({ errors }) {
    const [open, setOpen] = useState(false);
    return (
      <>
        {errors.length > 0 && (
          <button 
            onClick={() => setOpen(true)}
            style={{position: 'fixed', bottom: 20, right: 20, background: 'orange', borderRadius: '50%', width: 40, height: 40}}
          >⚠️</button>
        )}
        {open && <ErrorPanel errors={errors} onClose={() => setOpen(false)} />}
      </>
    );
  }
  