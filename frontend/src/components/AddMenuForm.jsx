import { useState } from "react"

function AddMenuForm({ initialValues, onSubmit, onCancel, submitLabel }) {
  const [form, setForm] = useState({
    name: initialValues?.name || "",
    price: initialValues?.price?.toString() || "",
    image_url: initialValues?.image_url || "",
  })

  const handleSubmit = async (e) => {
    e.preventDefault()
    await onSubmit({
      name: form.name.trim(),
      price: Number(form.price),
      image_url: form.image_url.trim(),
    })

    if (!initialValues) {
      setForm({ name: "", price: "", image_url: "" })
    }
  }

  return (
    <form className="menu-form" onSubmit={handleSubmit}>
      <input
        required
        placeholder="Food name"
        value={form.name}
        onChange={(e) => setForm((prev) => ({ ...prev, name: e.target.value }))}
      />
      <input
        required
        type="number"
        min="1"
        step="0.01"
        placeholder="Price"
        value={form.price}
        onChange={(e) => setForm((prev) => ({ ...prev, price: e.target.value }))}
      />
      <input
        placeholder="Image URL (optional)"
        value={form.image_url}
        onChange={(e) => setForm((prev) => ({ ...prev, image_url: e.target.value }))}
      />
      <button type="submit">{submitLabel || "Save"}</button>
      {onCancel ? (
        <button type="button" className="ghost" onClick={onCancel}>
          Cancel
        </button>
      ) : null}
    </form>
  )
}

export default AddMenuForm
