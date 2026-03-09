import { useState } from "react"

function AddMenu({ refresh }) {

  const [name,setName] = useState("")
  const [price,setPrice] = useState("")

  const createMenu = async () => {

    await fetch("http://localhost:3002/menu",{
      method:"POST",
      headers:{
        "Content-Type":"application/json"
      },
      body: JSON.stringify({
        name,
        price: parseFloat(price)
      })
    })

    setName("")
    setPrice("")

    refresh()
  }

  return (

    <div>

      <h2>Add Menu</h2>

      <input
        placeholder="name"
        value={name}
        onChange={(e)=>setName(e.target.value)}
      />

      <input
        placeholder="price"
        value={price}
        onChange={(e)=>setPrice(e.target.value)}
      />

      <button onClick={createMenu}>
        Add
      </button>

    </div>

  )
}

export default AddMenu