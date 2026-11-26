import java.util.*;
import java.time.*;

enum VehicleType { BIKE, CAR, TRUCK }

class Vehicle {
    VehicleType type;
    String regNo;

    public Vehicle(VehicleType type, String regNo) {
        this.type = type;
        this.regNo = regNo;
    }
}

class Slot {
    boolean occupied;
    LocalDateTime startTime;
    LocalDateTime endTime;
    double rate; // per second

    Slot(double rate) {
        this.occupied = false;
        this.rate = rate;
    }

    public void park() {
        occupied = true;
        startTime = LocalDateTime.now();
    }

    public void unpark() {
        occupied = false;
        endTime = LocalDateTime.now();
    }
}

class Ticket {
    int slotNumber;
    Vehicle vehicle;

    public Ticket(int slotNumber, Vehicle vehicle) {
        this.slotNumber = slotNumber;
        this.vehicle = vehicle;
    }
}

class ParkingLot {
    Map<VehicleType, List<Slot>> slots;
    public ParkingLot(int bikeSlots, int carSlots, int truckSlots) {
        slots = new HashMap<>();
        slots.put(VehicleType.BIKE, createSlots(bikeSlots, 1.0));
        slots.put(VehicleType.CAR, createSlots(carSlots, 2.0));
        slots.put(VehicleType.TRUCK, createSlots(truckSlots, 3.0));
    }

    private List<Slot> createSlots(int count, double rate) {
        List<Slot> result = new ArrayList<>(count);
        for (int i = 0; i < count; i++) {
            result.add(new Slot(rate));
        }
        return result;
    }
    
    public Ticket park(Vehicle vehicle) {
        for (int i = 0; i < slots.get(vehicle.type).size(); i++) {
            Slot slot = slots.get(vehicle.type).get(i);
            if (!slot.occupied) {
                slot.park();
                return new Ticket(i, vehicle);
            }
        }
        // No available slots
        return null;
    }
    public double unpark(Ticket ticket) {
        Slot slot = slots.get(ticket.vehicle.type).get(ticket.slotNumber);
        slot.unpark();
        Duration duration = Duration.between(slot.startTime, slot.endTime);
        double seconds = duration.toSeconds();
        return seconds * slot.rate;
    }

    public void displayAvailableSlots() {
        for (VehicleType type : slots.keySet()) {
            long availableCount = slots.get(type).stream().filter(s -> !s.occupied).count();
            System.out.println(type + " available slots: " + availableCount);
        }
    }
}

public class Main {
    public static void main(String[] args){
        ParkingLot parkingLot = new ParkingLot(4, 4, 2);
        Vehicle bike = new Vehicle(VehicleType.BIKE, "BIKE123");

        Ticket ticket = parkingLot.park(bike);
        parkingLot.displayAvailableSlots();

        if (ticket != null) {
            System.out.println("Bike parked at slot: " + ticket.slotNumber);
            // Simulate some time passing
            try { Thread.sleep(3000); } catch (InterruptedException e) {}
            double charge = parkingLot.unpark(ticket);
            System.out.println("Bike unparked. Charge: $" + charge);
        } else {
            System.out.println("No available slots for bike.");
        }
    }
}
