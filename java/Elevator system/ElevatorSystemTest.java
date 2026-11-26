public class ElevatorSystemTest {
    public static void main(String[] args) {
        testMaintenanceFinishesStopsAndStaysUnavailable();
        testMovingElevatorDoesNotTakeOppositeSideHallCall();
        testHallCallIsNotQueuedWhenAllElevatorsAreUnderMaintenance();
    }

    private static void testMaintenanceFinishesStopsAndStaysUnavailable() {
        Elevator elevator = new Elevator(1);
        elevator.dispatch(2);
        elevator.setMaintenance(true);

        for (int i = 0; i < 10 && (!elevator.minq.isEmpty() || !elevator.maxq.isEmpty()); i++) {
            elevator.move();
        }

        report(
            "maintenance elevator remains unavailable after finishing stops",
            elevator.state == State.MAINTENANCE,
            "state=" + elevator.state
        );
    }

    private static void testMovingElevatorDoesNotTakeOppositeSideHallCall() {
        ElevatorSystem system = new ElevatorSystem(1);
        Elevator elevator = system.elevators.get(0);
        elevator.currentFloor = 5;
        elevator.direction = Direction.UP;
        elevator.state = State.MOVING;

        system.request(2, Direction.UP);

        report(
            "up-moving elevator does not accept an up call behind it",
            !elevator.minq.contains(2) && !elevator.maxq.contains(2),
            "upStops=" + elevator.minq + ", downStops=" + elevator.maxq
        );
    }

    private static void testHallCallIsNotQueuedWhenAllElevatorsAreUnderMaintenance() {
        ElevatorSystem system = new ElevatorSystem(1);
        Elevator elevator = system.elevators.get(0);
        elevator.setMaintenance(true);

        system.request(4, Direction.UP);

        report(
            "hall call is not queued when all elevators are under maintenance",
            elevator.state == State.MAINTENANCE
                && elevator.minq.isEmpty()
                && elevator.maxq.isEmpty(),
            "state=" + elevator.state
                + ", upStops=" + elevator.minq
                + ", downStops=" + elevator.maxq
        );
    }

    private static void report(String name, boolean passed, String details) {
        System.out.println((passed ? "PASS" : "FAIL") + ": " + name + " (" + details + ")");
    }
}
