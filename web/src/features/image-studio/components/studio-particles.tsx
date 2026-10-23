/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { useEffect, useRef } from 'react'

export function StudioParticles() {
  const ref = useRef<HTMLCanvasElement>(null)

  useEffect(() => {
    const canvas = ref.current
    if (!canvas || typeof ResizeObserver === 'undefined') return
    const context = canvas.getContext('2d')
    if (!context) return
    const motion = window.matchMedia('(prefers-reduced-motion: reduce)')
    const particles = Array.from({ length: 96 }, () => ({
      x: Math.random(),
      y: Math.random(),
      radius: 0.6 + Math.random() * 0.8,
      speed: 7 + Math.random() * 13,
      phase: Math.random() * Math.PI * 2,
    }))
    let width = 0
    let height = 0
    let frame = 0
    let previous = 0
    let elapsed = 0

    const paint = () => {
      if (!width || !height) return
      context.clearRect(0, 0, width, height)
      for (
        let index = 0;
        index < (width < 640 ? 32 : particles.length);
        index++
      ) {
        const particle = particles[index]
        const x =
          ((particle.x * width + elapsed * particle.speed * 0.3) % width) +
          Math.sin(elapsed * 0.4 + particle.phase) * 8
        const y =
          (((particle.y * height - elapsed * particle.speed) % height) +
            height) %
          height
        const alpha =
          0.22 + (Math.sin(elapsed * 1.2 + particle.phase) + 1) * 0.26
        context.fillStyle = `rgba(209, 214, 255, ${alpha})`
        context.beginPath()
        context.arc(x, y, particle.radius, 0, Math.PI * 2)
        context.fill()
        if (index % 16 === 0 && !motion.matches) {
          context.strokeStyle = `rgba(177, 180, 255, ${alpha * 0.2})`
          context.lineWidth = 0.6
          context.beginPath()
          context.moveTo(x, y)
          context.lineTo(x - 3, y + 10)
          context.stroke()
        }
      }
    }
    const tick = (now: number) => {
      frame = requestAnimationFrame(tick)
      if (previous && now - previous < 1000 / 30) return
      elapsed += previous ? Math.min((now - previous) / 1000, 0.1) : 0
      previous = now
      paint()
    }
    const resume = () => {
      cancelAnimationFrame(frame)
      previous = 0
      paint()
      if (!document.hidden && !motion.matches) {
        frame = requestAnimationFrame(tick)
      }
    }
    const observer = new ResizeObserver(() => {
      const bounds = canvas.getBoundingClientRect()
      width = bounds.width
      height = bounds.height
      const scale = Math.min(window.devicePixelRatio || 1, 1.5)
      canvas.width = Math.round(width * scale)
      canvas.height = Math.round(height * scale)
      context.setTransform(scale, 0, 0, scale, 0, 0)
      resume()
    })
    observer.observe(canvas)
    document.addEventListener('visibilitychange', resume)
    motion.addEventListener('change', resume)
    return () => {
      cancelAnimationFrame(frame)
      observer.disconnect()
      document.removeEventListener('visibilitychange', resume)
      motion.removeEventListener('change', resume)
    }
  }, [])

  return <canvas ref={ref} aria-hidden='true' className='studio-particles' />
}
