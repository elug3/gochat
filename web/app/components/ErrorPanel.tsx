function ErrorPanel({ errors, onClose }) {
    return (
      <div style={{position: 'fixed', bottom: 60, right: 20, width: 300, maxHeight: 400, overflow: 'auto', background: '#fff', border: '1px solid #ccc'}}>
        <button onClick={onClose}>Close</button>
        <h4>Errors ({errors.length})</h4>
        <ul>
          {errors.map((err, i) => (
            <li key={i}>
              <strong>{err.type}</strong>: {err.message} <br/>
              <small>{err.file}:{err.line}</small>
              <pre style={{whiteSpace: 'pre-wrap'}}>{err.stack}</pre>
            </li>
          ))}
        </ul>
        <button onClick={reportErrors}>Report</button>
      </div>
    );
} 